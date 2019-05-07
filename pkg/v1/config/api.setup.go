package config

import (
	"github.com/Atluss/FileServerWithMQ/pkg/v1"
	"github.com/gorilla/mux"
	"github.com/nats-io/go-nats"
	"time"
)

// NewApiSetup return config with nats and router
func NewApiSetup(settings string) *Setup {

	cnf, err := Config(settings)
	v1.FailOnError(err, "error config file")

	set, err := newSetup(cnf)
	v1.FailOnError(err, "error setup")

	return set
}

func newSetup(cnf *config) (*Setup, error) {

	set := Setup{}

	if err := cnf.validate(); err != nil {
		return &set, err
	}

	set.Config = cnf

	if err := set.natsConnection(); err != nil {
		return &set, err
	}

	set.Route = mux.NewRouter().StrictSlash(true)

	return &set, nil
}

// Setup main setup api struct
type Setup struct {
	Config *config     // api setting
	Nats   *nats.Conn  // nats
	Route  *mux.Router // mux frontend
}

// natsConnection setup nats
func (obj *Setup) natsConnection() error {

	var err error

	if obj.Nats, err = nats.Connect(obj.Config.Nats.Address[0].Address, nats.MaxReconnects(-1), nats.ReconnectWait(time.Second*5)); err != nil {
		return err
	}

	return nil
}
