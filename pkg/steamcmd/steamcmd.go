package steamcmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Cehir/steam-workshop-downloader/pkg/config"
	"github.com/Cehir/steam-workshop-downloader/pkg/path"
	logger "github.com/sirupsen/logrus"
	"os/exec"
	"path/filepath"
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	var cmdArgs []string
	// set login credentials
	cmdArgs = append(cmdArgs, s.cfg.Steam.Login.CmdArgs()...)
	// add +workshop_download_item <appid> <modid> <install dir> <validate>
	cmdArgs = append(cmdArgs, s.cfg.Apps.CmdArgs()...)
	// quit after login
	cmdArgs = append(cmdArgs, "+quit")

	logger.WithField("args", cmdArgs).Debug("steamcmd args")

	cmd := exec.CommandContext(ctx, s.cfg.Steam.Cmd, cmdArgs...)

	// done channel
	done := make(chan error, 1)

	// init scanner
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanLines)

	appDestination := s.cfg.Apps.Destinations()

	// start scanner
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			logger.Debug(text)
			if downloadFolder := extractPathRegex.FindStringSubmatch(text); downloadFolder != nil {
				// extract workshop id from path
				if appID := appIDRegex.FindStringSubmatch(downloadFolder[1]); appID != nil {
					f := filepath.Join(downloadFolder[1], "mods")
					err := path.CopyDir(f, appDestination[appID[1]])

					if err != nil {
						logger.WithError(err).
							WithField("workshop_id", appID[1]).
							WithField("source", f).
							WithField("destination", appDestination[appID[1]]).
							Error("failed to copy mod")
						return
					}
				}
			}
		}
	}()

	// start steamcmd
	if err := cmd.Start(); err != nil {
		logger.WithError(err).Error("failed to run steamcmd")
	}

	go func() {
		done <- cmd.Wait()
	}()

	// wait for steamcmd to finish
	select {
	case <-ctx.Done():
		// kill steamcmd if context is done
		if err := cmd.Process.Kill(); err != nil {
			logger.WithError(err).Error("failed to kill steamcmd")
		}
		return ctx.Err()
	case err := <-done:
		return err
	}
}
