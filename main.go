package main

import (
	"context"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/gen2brain/beeep"
	"github.com/sirupsen/logrus"
)

//go:embed logo_small.png
var logo []byte

//go:embed VERSION
var version string

var log = logrus.StandardLogger()

func main() {
	if err := main_(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func main_() error {
	generate := flag.Bool("g", false, "generate config.toml in the working directory and exit")
	flag.Parse()

	if *generate {
		err := generateConfig()
		if err != nil {
			return fmt.Errorf("failed to generate config.toml: %s", err)
		}
		log.Infof("generated config.toml")
		return nil
	}

	cfg := loadAndParseConfig()

	if cfg.Verbose {
		log.SetLevel(logrus.DebugLevel)
	}

	if !strings.Contains(cfg.MessageOnConnection, "%s") {
		return errors.New("'MessageOnConnection' must contain '%s'")
	}

	if !strings.Contains(cfg.MessageOnDisconnection, "%s") {
		return errors.New("'MessageOnDisconnection' must contain '%s'")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	w, err := newWatcher()
	if err != nil {
		return fmt.Errorf("failed to create a new watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go w.Start(ctx, wg)
	defer wg.Wait()
	defer cancel()

	logoPath, err := prepareLogo()
	if err != nil {
		log.Errorf("failed to prepare logo in a temporary directory: %w", err)
	}
	defer os.Remove(logoPath)

	log.Infof("started")

	for {
		select {
		case ev := <-w.Events():
			switch ev.Type {
			case EventConnect:
				if !cfg.DetectConnection {
					continue
				}
				err = beeep.Notify(cfg.Title, fmt.Sprintf(cfg.MessageOnConnection, ev.Path), logoPath)
				if err != nil {
					return fmt.Errorf("failed to notify: %s", err)
				}
			case EventDisconnect:
				if !cfg.DetectDisconnection {
					continue
				}
				err = beeep.Notify(cfg.Title, fmt.Sprintf(cfg.MessageOnDisconnection, ev.Path), logoPath)
				if err != nil {
					return fmt.Errorf("failed to notify: %s", err)
				}
			case EventError:
				return fmt.Errorf("failed to watch ports: %s", ev.Error)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func prepareLogo() (string, error) {
	lf, err := os.CreateTemp("", "port-notifier-logo.png")
	if err != nil {
		return "", fmt.Errorf("failed to create a logo file: %w", err)
	}
	defer lf.Close()

	log.Debugf("logo file: %s", lf.Name())

	_, err = lf.Write(logo)
	if err != nil {
		return "", fmt.Errorf("failed to write a logo: %w", err)
	}
	return lf.Name(), nil
}
