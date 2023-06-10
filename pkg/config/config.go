package config

import (
	"encoding/json"
	"fmt"
	"github.com/Cehir/steam-workshop-downloader/pkg/path"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"strings"
)

var (
	Validator = validator.New()
	Path      = path.NewPath()
)

func init() {
	Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}
		return name
	})
}

type Apps []*App

func (a *Apps) String() string {
	if a == nil {
		return ""
	}
	return strings.Join(a.Strings(), ",")
}

func (a *Apps) Strings() []string {
	if a == nil {
		return nil
	}
	var s []string
	for _, app := range *a {
		s = append(s, app.String())
	}
	return s
}

func (a *Apps) CmdArgs() []string {
	if a == nil {
		return nil
	}
	var s []string
	for _, app := range *a {
		for _, mod := range app.Mods {
			s = append(s, "+workshop_download_item", app.AppID, mod.WorkshopID)
		}
	}
	return s
}

// Destinations returns a map of appID to destination path
func (a *Apps) Destinations() map[string]string {
	if a == nil {
		return nil
	}
	m := make(map[string]string, len(*a))
	for _, app := range *a {
		m[app.AppID] = app.Path
	}
	return m
}

type Config struct {
	Steam Steam `json:"steam" mapstructure:"steam" validate:"required"`                        // Steam config
	Apps  Apps  `json:"apps,omitempty" mapstructure:"apps" validate:"omitempty,dive,required"` // List of games with mods to download
}

type Steam struct {
	Login Login  `json:"login" mapstructure:"login" validate:"required"`  // Login credentials
	Cmd   string `json:"cmd" mapstructure:"cmd" validate:"required,file"` // SteamCMD path e.g. /usr/bin/steamcmd
}

// Validate validates the config
func (c *Config) Validate() error {
	cp, err := Path.Absolute(c.Steam.Cmd)
	if err != nil {
		return err
	}
	c.Steam.Cmd = cp

	return Validator.Struct(c)
}

type ModPath struct {
	AppName string
	AppPath string
	ModName string
}

func (c *Config) ExportPath(appID, modId string) (ModPath, error) {
	if c == nil {
		return ModPath{}, fmt.Errorf("config is nil")
	}

	for _, app := range c.Apps {
		if app.AppID == appID {
			for _, mod := range app.Mods {
				if mod.WorkshopID == modId {
					return ModPath{
						AppName: app.Name,
						AppPath: app.Path,
						ModName: mod.Name,
					}, nil
				}
			}
		}
	}

	return ModPath{}, fmt.Errorf("mod %s not found for app %s", modId, appID)
}

func (c *Config) PrintYAML() error {
	return yaml.NewEncoder(os.Stdout).Encode(&c)
}

func (c *Config) PrintJSON() error {
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	return e.Encode(&c)
}

type Login struct {
	Username string `json:"username" mapstructure:"username" validate:"required"` // Username
	Password string `json:"password" mapstructure:"password"`                     // Password
}

// String returns the username and masked password if set
func (l *Login) String() string {
	if l == nil {
		return ""
	}
	if l.Password != "" {
		return fmt.Sprintf("%s:***", l.Username)
	}
	return l.Username
}

// Validate validates the login
func (l *Login) Validate() error {
	return Validator.Struct(l)
}

// Login for steamcmd
func (l *Login) CmdArgs() []string {
	if l == nil {
		return nil
	}
	if l.Password != "" {
		return []string{"+login", l.Username, l.Password}
	}
	return []string{"+login", l.Username}
}

type App struct {
	Name  string `json:"name" mapstructure:"name"`                                              // Name of the game
	AppID string `json:"id" mapstructure:"id" validate:"required"`                              // Steam App ID
	Path  string `json:"path,omitempty" mapstructure:"path" validate:"required,dir"`            // Path to the mod directory
	Mods  []*Mod `json:"mods,omitempty" mapstructure:"mods" validate:"omitempty,dive,required"` // List of mods to download for the game
}

func (a *App) String() string {
	if a == nil {
		return ""
	}
	return fmt.Sprintf("%s (%s)", a.Name, a.AppID)
}

type Mod struct {
	Name       string `json:"name,omitempty" mapstructure:"name"`       // Name of the mod
	WorkshopID string `json:"id" mapstructure:"id" validate:"required"` // Steam Workshop ID
}
