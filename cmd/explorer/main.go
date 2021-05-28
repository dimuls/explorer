package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"explorer"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	log := logrus.WithField("subsystem", "main")

	configPath := os.Getenv("EXPLORER_CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}

	log.WithField("config_path", configPath).Info("loading config")
	configYAML, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.WithError(err).Fatal("failed to read config")
	}

	var ec explorer.ExplorerConfig

	err = yaml.Unmarshal(configYAML, &ec)
	if err != nil {
		log.WithError(err).Fatal("failed to parse config")
	}

	log.Info("creating and starting explorer")
	e, err := explorer.New(ec)
	if err != nil {
		log.WithError(err).Fatal("failed to create explorer")
	}

	time.Sleep(200 * time.Second)

	log.Info("explorer started")

	exit := make(chan os.Signal)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	log.WithField("signal", <-exit).Info("exit signal received")

	log.Info("closing explorer")
	e.Close()

	log.Info("explorer closed, exiting")
}
