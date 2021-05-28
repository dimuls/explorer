package explorer

import (
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"explorer/hf"
	"explorer/pg"
)

type ExplorerConfig struct {
	StorageDSN string               `yaml:"storage_dsn"`
	Processors []hf.ProcessorConfig `yaml:"processors"`
}

type Explorer struct {
	processors []*hf.Processor
	log        *logrus.Entry
}

func New(c ExplorerConfig) (e *Explorer, err error) {

	log := logrus.WithField("subsystem", "explorer")

	s, err := pg.NewStorage(c.StorageDSN)
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

	return &Explorer{
		processors: ps,
		log:        log,
	}, nil
}

func (e *Explorer) Close() {
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
