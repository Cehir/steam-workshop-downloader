package output

import (
	"errors"
	"github.com/Cehir/steam-workshop-downloader/pkg/config"
)

type Output string

const (
	YAML Output = "yaml"
	JSON Output = "json"
)

var (
	InvalidFormatErr = errors.New(`invalid output format, must be "yaml" or "json"`)
)

// String returns the string representation of the output
// it is used to implement the flag.Value interface
func (o *Output) String() string {
	return string(*o)
}

// Set sets the output to the given value
// it is used to implement the flag.Value interface
func (o *Output) Set(v string) error {
	switch v {
	case "yaml", "json":
		*o = Output(v)
		return nil
	default:
		return InvalidFormatErr
	}
}

// Type returns the type of the output
// it is used to implement the flag.Value interface
func (o *Output) Type() string {
	return "output"
}

func (o *Output) Print(cfg *config.Config) error {
	switch *o {
	case YAML:
		return cfg.PrintYAML()
	case JSON:
		return cfg.PrintJSON()
	default:
		return InvalidFormatErr
	}
}
