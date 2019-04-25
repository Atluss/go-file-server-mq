package config

import (
	"encoding/json"
	"fmt"
	"github.com/Atluss/FileServerWithMQ/lib"
	"io/ioutil"
	"os"
	"strings"
)

// config load new config for API
func Config(path string) (*config, error) {

	conf := config{}

	if err := lib.CheckFileExist(path); err != nil {
		return &conf, err
	}

	conf.FilePath = path

	if err := conf.load(); err != nil {
		return &conf, err
	}

	return &conf, nil
}

// natsConfig main load nats config
type natsConfig struct {
	Version         string           `json:"Version"`         // nats version
	ReconnectedWait int              `json:"ReconnectedWait"` // nats ReconnectedWait
	Address         []natsCnfAddress `json:"Address"`         // list off nats servers
}

// natsCnfAddress nats add array element
type natsCnfAddress struct {
	Host    string `json:"Host"`
	Port    string `json:"Port"`
	Address string `json:"Address"`
}

// config main
type config struct {
	Name     string     `json:"Name"`    // API name
	Version  string     `json:"Version"` // API version
	Host     string     `json:"Host"`
	Port     string     `json:"Port"`
	FilePath string     `json:"FilePath"` // path to Json settings file
	Nats     natsConfig `json:"Nats"`
}

// load all settings
func (obj *config) load() error {

	jsonSet, err := os.Open(obj.FilePath)

	defer func() {
		// defer and handle close error
		lib.LogOnError(jsonSet.Close(), "warning: Can't close json settings file.")
	}()

	if !lib.LogOnError(err, "Can't open config file") {
		return err
	}

	bytesVal, _ := ioutil.ReadAll(jsonSet)
	err = json.Unmarshal(bytesVal, &obj)

	if !lib.LogOnError(err, "Can't unmarshal json file") {
		return err
	}

	return obj.validate()
}

// validate it
func (obj *config) validate() error {

	if obj.Name == "" {
		return fmt.Errorf("config miss name")
	}

	if obj.Version == "" {
		return fmt.Errorf("config miss version")
	}

	if obj.Host == "" {
		return fmt.Errorf("config miss host")
	}

	if obj.Port == "" {
		return fmt.Errorf("config miss port")
	}

	if obj.Nats.Version == "" {
		return fmt.Errorf("config miss Nats version")
	}

	if obj.Nats.ReconnectedWait == 0 {
		return fmt.Errorf("config miss Nats ReconnectedWait")
	}

	if obj.Nats.Address[0].Host == "" {
		return fmt.Errorf("config miss Nats host")
	}

	if obj.Nats.Address[0].Port == "" {
		return fmt.Errorf("config miss Nats port")
	}

	obj.assemble()

	return nil
}

// assemble write Nats address
func (obj *config) assemble() {
	obj.Nats.Address[0].Address = fmt.Sprintf("nats://%s:%s", obj.Nats.Address[0].Host, obj.Nats.Address[0].Port)
}

// GetNatsAddresses that other way to get nats connections addresses for clusters purpose
func (obj *config) GetNatsAddresses() string {

	addresses := make([]string, len(obj.Nats.Address))

	for i, ns := range obj.Nats.Address {
		addresses[i] = fmt.Sprintf("%s:%s", ns.Host, ns.Port)
	}

	str := fmt.Sprintf("nats://%s", strings.Join(addresses, ", "))

	return str
}
