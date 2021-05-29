package explorer

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"explorer/hf"
	"explorer/pg"
)

type AppConfig struct {
	StorageDSN string               `yaml:"storage_dsn"`
	Processors []hf.ProcessorConfig `yaml:"processors"`
}

type App struct {
	processors []*hf.Processor
	log        *logrus.Entry
}

func New(c AppConfig) (e *App, err error) {

	log := logrus.WithField("subsystem", "explorer")

	s, err := pg.NewExplorer(c.StorageDSN)
	if err != nil {
		return nil, fmt.Errorf("create storage: %w", err)
	}

	defer func() {
		if err != nil {
			err := s.Close()
			if err != nil {
				log.WithError(err).Error("failed to close storage")
			}
		}
	}()

	err = s.Migrate()
	if err != nil {
		return nil, fmt.Errorf("migrate storage: %w", err)
	}

	var ps []*hf.Processor
	for _, pc := range c.Processors {

		log.WithField("channel_id", pc.ChannelID).
			Info("creating and starting processor")

		p, pErr := hf.NewProcessor(pc, s)
		if pErr != nil {
			return nil, fmt.Errorf(
				"create processor for channel_id=`%s`: %w",
				pc.ChannelID, pErr)
		}

		defer func(p *hf.Processor) {
			if err != nil {
				p.Close()
			}
		}(p)

		ps = append(ps, p)
	}

	log.Info("explorer started")

	return &App{
		processors: ps,
		log:        log,
	}, nil
}

func (e *App) Close() {
	var wg sync.WaitGroup

	for _, p := range e.processors {
		wg.Add(1)
		go func(p *hf.Processor) {
			wg.Done()
			p.Close()
		}(p)
	}
	wg.Wait()
}

func Run() {
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

	var ec AppConfig

	err = yaml.Unmarshal(configYAML, &ec)
	if err != nil {
		log.WithError(err).Fatal("failed to parse config")
	}

	log.Info("creating and starting explorer")
	e, err := New(ec)
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
