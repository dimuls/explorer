package migrations

import (
	"embed"
	"net/http"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
)

//go:embed *.sql
var migrations embed.FS

type embedFSDriver struct {
	httpfs.PartialDriver
}

func init() {
	source.Register("embed", &embedFSDriver{})
}

func (d *embedFSDriver) Open(rawURL string) (source.Driver, error) {
	err := d.PartialDriver.Init(http.FS(migrations), ".")
	if err != nil {
		return nil, err
	}

	return d, nil
}
