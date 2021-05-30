package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	"explorer"
	"explorer/hf"
	_ "explorer/pg/migrations"
	"explorer/ui"
)

type ExplorerConfig struct {
	Listen struct {
		GRPC string `yaml:"grpc"`
		HTTP string `yaml:"http"`
	} `yaml:"listen"`
	DSN        string               `yaml:"dsn"`
	Processors []hf.ProcessorConfig `yaml:"processors"`
}

type Explorer struct {
	config      ExplorerConfig
	sqlDB       *sql.DB
	db          *goqu.Database
	processors  []*hf.Processor
	tcpListener net.Listener
	grpcServer  *grpc.Server
	httpServer  *http.Server
	log         *logrus.Entry

	explorer.UnimplementedExplorerServer
}

func NewExplorer(c ExplorerConfig) (*Explorer, error) {
	return &Explorer{
		config: c,
		log: logrus.WithFields(logrus.Fields{
			"subsystem": "pg_explorer",
		}),
	}, nil
}

func (e *Explorer) Migrate() error {

	sqlDB, err := sql.Open(postgresDialect, e.config.DSN)
	if err != nil {
		return fmt.Errorf("open DB: %w", err)
	}

	d, err := postgres.WithInstance(sqlDB, &postgres.Config{
		MigrationsTable: "migration",
	})
	if err != nil {
		return fmt.Errorf("create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"embed://", postgresDialect, d)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}

func (e *Explorer) Run() (err error) {

	e.sqlDB, err = sql.Open(postgresDialect, e.config.DSN)
	if err != nil {
		return fmt.Errorf("open DB: %w", err)
	}

	defer func() {
		if err != nil {
			err := e.sqlDB.Close()
			if err != nil {
				e.log.WithError(err).Error("failed to close DB")
			}
		}
	}()

	e.db = goqu.New(postgresDialect, e.sqlDB)

	err = e.Migrate()
	if err != nil {
		return fmt.Errorf("migrate DB: %w", err)
	}

	for _, pc := range e.config.Processors {

		e.log.WithField("channel_id", pc.ChannelID).
			Info("creating and starting processor")

		p, pErr := hf.NewProcessor(pc, e)
		if pErr != nil {
			return fmt.Errorf(
				"create processor for channel_id=`%s`: %w",
				pc.ChannelID, pErr)
		}

		defer func(p *hf.Processor) {
			if err != nil {
				p.Close()
			}
		}(p)

		e.processors = append(e.processors, p)
	}

	e.grpcServer = grpc.NewServer(grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			grpc_logrus.UnaryServerInterceptor(e.log),
			e.UnaryAuthInterceptor,
		),
	))

	explorer.RegisterExplorerServer(e.grpcServer, e)

	e.tcpListener, err = net.Listen("tcp", e.config.Listen.GRPC)
	if err != nil {
		logrus.WithError(err).Error("failed to create tcp_listener")
		return
	}

	go func() {
		for {
			err := e.grpcServer.Serve(e.tcpListener)
			if err != nil {
				e.log.WithError(err).Error("failed to start GRPC server")
			}
			time.Sleep(3 * time.Second)
		}
	}()

	explorerAPIMux := runtime.NewServeMux()

	err = explorer.RegisterExplorerHandlerFromEndpoint(
		context.Background(), explorerAPIMux, e.tcpListener.Addr().String(),
		[]grpc.DialOption{grpc.WithInsecure()})
	if err != nil {
		return fmt.Errorf("failed to init serve mux: %w", err)
	}

	httpRouter := mux.NewRouter()

	httpRouter.PathPrefix("/api/").Handler(explorerAPIMux)

	swaggerFile, err := explorer.FS.Open(explorer.SwaggerFile)
	if err != nil {
		return fmt.Errorf("failed to open swagger file: %w", err)
	}

	swaggerJSON, err := ioutil.ReadAll(swaggerFile)
	swaggerFile.Close()
	if err != nil {
		return fmt.Errorf("failed to read swagger file: %w", err)
	}

	httpRouter.Path("/swagger.json").Methods(http.MethodGet).
		HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write(swaggerJSON)
		})

	indexFile, err := ui.FS.Open(path.Join(ui.Prefix, "index.html"))
	if err != nil {
		logrus.WithError(err).Error("failed to open index file")
		return
	}

	indexHTML, err := ioutil.ReadAll(indexFile)
	indexFile.Close()
	if err != nil {
		logrus.WithError(err).Error("failed to read index file")
		return
	}

	httpRouter.PathPrefix("/").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			staticPath := path.Join(ui.Prefix, r.URL.Path)

			content, err := ui.FS.ReadFile(staticPath)
			if err != nil {
				w.Write(indexHTML)
				return
			}

			w.Header().Set("Content-Type",
				mime.TypeByExtension(path.Ext(staticPath)))
			w.Write(content)
		})

	e.httpServer = &http.Server{
		Addr:    e.config.Listen.HTTP,
		Handler: httpRouter,
	}

	go func() {
		for {
			err := e.httpServer.ListenAndServe()
			if err != nil {
				if err == http.ErrServerClosed {
					break
				}
				e.log.WithError(err).Error("failed to start HTTP server")
			}
			time.Sleep(3 * time.Second)
		}
	}()

	return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
	defer cancel()

	wg.Add(1)
	go func() {
		wg.Done()
		err := e.httpServer.Shutdown(ctx)
		if err != nil {
			e.log.WithError(err).Errorf("failed to stop HTTP server: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		wg.Done()
		e.grpcServer.GracefulStop()
	}()

	wg.Wait()

	err := e.tcpListener.Close()
	if err != nil {
		e.log.WithError(err).Error("failed to close TCP listener: %w", err)
	}

	err = e.sqlDB.Close()
	if err != nil {
		e.log.WithError(err).Error("failed to close DB connection: %w", err)
	}

	return
}

func RunExplorer() {
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

	var ec ExplorerConfig

	err = yaml.Unmarshal(configYAML, &ec)
	if err != nil {
		log.WithError(err).Fatal("failed to parse config")
	}

	log.Info("creating and starting explorer")

	e, err := NewExplorer(ec)
	if err != nil {
		log.WithError(err).Fatal("failed to create explorer")
	}

	err = e.Migrate()
	if err != nil {
		log.WithError(err).Fatal("failed to migrate explorer DB")
	}

	err = e.Run()
	if err != nil {
		log.WithError(err).Fatal("failed to run explorer")
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
