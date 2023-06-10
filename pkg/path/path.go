package path

import (
	logger "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Path struct{}

func NewPath() *Path {
	return &Path{}
}

// Absolute returns the absolute path of the given path
func (p *Path) Absolute(path string) (string, error) {
	if path == "$HOME" || path == "~" || path == "%userprofile%" {
		return os.UserHomeDir()
	}

	if strings.HasPrefix(path, "$HOME") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = home + path[5:]
	}

	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = home + path[1:]
	}

	if strings.HasPrefix(path, "%userprofile%") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = home + path[13:]
	}

	inPath := os.ExpandEnv(path)

	if filepath.IsAbs(inPath) {
		return filepath.Clean(inPath), nil
	}

	fp, err := filepath.Abs(inPath)
	if err != nil {
		return "", err
	}

	return filepath.Clean(fp), nil
}

// CopyDir copies the content of src to dst. src should be a full path.
func CopyDir(src, dst string) error {

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// copy to this path
		outpath := filepath.Join(dst, strings.TrimPrefix(path, src))
		logger.WithFields(logger.Fields{
			"from": path,
			"to":   outpath,
		}).Debug("copying")

		if info.IsDir() {
			err := os.MkdirAll(outpath, info.Mode())
			if err != nil {
				return err
			}
			return nil // means recursive
		}

		// handle irregular files
		if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return os.Symlink(link, outpath)
			}
			return nil
		}

		// copy contents of regular file efficiently

		// open input
		in, _ := os.Open(path)
		if err != nil {
			return err
		}
		defer func(in *os.File) {
			_ = in.Close()
		}(in)

		// create output
		fh, err := os.Create(outpath)
		if err != nil {
			return err
		}
		defer func(fh *os.File) {
			_ = fh.Close()
		}(fh)

		// make it the same
		err = fh.Chmod(info.Mode())
		if err != nil {
			return err
		}

		// copy content
		_, err = io.Copy(fh, in)
		return err
	})
}
