package steamcmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Cehir/steam-workshop-downloader/pkg/config"
	"github.com/Cehir/steam-workshop-downloader/pkg/path"
	logger "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type SteamCmd struct {
	cfg *config.Config
}

func NewSteamCmd(cfg *config.Config) *SteamCmd {
	return &SteamCmd{
		cfg: cfg,
	}
}

func (s *SteamCmd) Download() error {
	errch := make(chan error, 2)
	downloadChan := make(chan download)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.cfg.Steam.Cmd)

	// get pipes
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.WithError(err).Error("failed to get stderr pipe")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.WithError(err).Error("failed to get stdout pipe")
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.WithError(err).Error("failed to get stdin pipe")
	}

	// start steamcmd
	if err := cmd.Start(); err != nil {
		logger.WithError(err).Error("failed to start steamcmd")
	}

	// read stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			logger.WithField("line", line).Debug("steamcmd error")
		}
	}()

	// read stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		fmt.Println("")
		for scanner.Scan() {
			line := scanner.Text()
			logger.WithField("line", line).Debug("steamcmd output")

			switch line {
			case "Loading Steam API...OK":
				s.loginAfterApiLoad(stdin, errch)
			case fmt.Sprintf("Steam>Logging in user '%s' to Steam Public...FAILED (Invalid Password)", s.cfg.Steam.Login.Username):
				s.quitOnFailedLogin(stdin, errch)
			case "Waiting for user info...OK":
				s.downloadModsAfterLogin(stdin, errch)
			default:
				checkForSuccessFullDownload(line, downloadChan)
			}
		}
	}()

	// wait for steamcmd to finish
	go func() {
		errch <- cmd.Wait()
	}()

	// wait for downloads to finish
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Error("steamcmd timed out")
				return
			case d := <-downloadChan:
				from := d.path
				to, err := s.cfg.ExportPath(d.app, d.mod)
				if err != nil {
					errch <- err
					return
				}

				err = path.CopyDir(from, to.AppPath)
				if err != nil {
					logger.WithFields(logger.Fields{
						"app":  to.AppName,
						"mod":  to.ModName,
						"from": from,
						"to":   to.AppPath,
					}).WithError(err).Error("failed to copy")
					errch <- err
					return
				}

				logger.WithFields(logger.Fields{
					"app": to.AppName,
					"mod": to.ModName,
				}).Info("copy mod to export path complete")
			}
		}
	}()

	select {
	case <-ctx.Done():
		logger.Error("steamcmd timed out")
		return ctx.Err()
	case err := <-errch:
		if err != nil {
			return err
		}
	}
	return err
}

func (s *SteamCmd) quitOnFailedLogin(stdin io.WriteCloser, errch chan error) {
	logger.WithField("username", s.cfg.Steam.Login.Username).Error("Steam login failed")
	// log error and quit on failed login
	if _, err := s.quit(stdin); err != nil {
		errch <- err
		return
	}
	errch <- fmt.Errorf("steam login failed")
}

// login writes the login arguments for steamcmd
func (s *SteamCmd) login(closer io.WriteCloser) (n int, err error) {
	return closer.Write([]byte(fmt.Sprintf("login %s %s\n", s.cfg.Steam.Login.Username, s.cfg.Steam.Login.Password)))
}

// quit writes the quit arguments for steamcmd
func (s *SteamCmd) quit(closer io.WriteCloser) (n int, err error) {
	return closer.Write([]byte("quit\n"))
}

// download writes the download arguments for steamcmd
func (s *SteamCmd) download(closer io.WriteCloser) (n int, err error) {
	var x int

	if s.cfg.Apps != nil {
		for _, app := range s.cfg.Apps {
			if app.Mods != nil {
				for _, mod := range app.Mods {
					x, err = closer.Write([]byte(fmt.Sprintf("workshop_download_item %s %s\n", app.AppID, mod.WorkshopID)))
					n += x
					if err != nil {
						return n, err
					}
				}
			}
		}
	}

	return n, nil
}

// loginAfterApiLoad writes the login arguments for steamcmd
func (s *SteamCmd) loginAfterApiLoad(stdin io.WriteCloser, errch chan error) {
	logger.Info("Steam API loaded")
	// send login after api load
	if _, err := s.login(stdin); err != nil {
		logger.WithError(err).Error("failed to send login to steamcmd")
		errch <- err
	}
}

// downloadModsAfterLogin writes the download and quit arguments for steamcmd
func (s *SteamCmd) downloadModsAfterLogin(stdin io.WriteCloser, errch chan error) {
	logger.Info("Steam user info loaded")
	// send download after login
	if _, err := s.download(stdin); err != nil {
		logger.WithError(err).Error("failed to send download to steamcmd")
		errch <- err
		return
	}

	// send quit after download
	if _, err := s.quit(stdin); err != nil {
		logger.WithError(err).Error("failed to send quit to steamcmd")
		errch <- err
		return
	}
}

var workshopIDRegex = regexp.MustCompile(`Success. Downloaded item (\d+) to "(.*)"`)
var results []string

const dirSep = string(os.PathSeparator)
const prefixN = len("Success. Downloaded item ")

type download struct {
	app  string
	mod  string
	path string
}

func checkForSuccessFullDownload(line string, downloadChan chan download) {
	results = workshopIDRegex.FindAllString(line, -1)

	// if not successfully downloaded, return
	if len(results) == 0 {
		return
	}

	//search for "to" and split on it
	n := strings.Split(results[0][prefixN:], " to ")

	//remove quotes from path
	p := n[1][1 : len(n[1])-1]

	//split path on directory separator
	dirSegments := strings.Split(p, dirSep)
	i := len(dirSegments)

	//workshop id is the last directory in the path
	workshopID := dirSegments[i-1]
	//app id is the second last directory in the path
	appID := dirSegments[i-2]

	downloadChan <- download{
		app:  appID,
		mod:  workshopID,
		path: p + dirSep + "mods",
	}
}
