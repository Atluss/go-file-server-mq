package config

import (
	"github.com/Atluss/FileServerWithMQ/pkg/v1"
	"log"
	"testing"
)

func TestConfig(t *testing.T) {

	path := "settings.json"

	cnf, err := Config(path)
	v1.FailOnError(err, "Test error")

	log.Printf("%+v", cnf)
	log.Printf("Name: %s", cnf.Name)
	log.Printf("Version: %s", cnf.Version)
	log.Printf("Nats version: %s", cnf.Nats.Version)
	log.Printf("Nats ReconnectedWait: %d", cnf.Nats.ReconnectedWait)
	log.Printf("Nats host: %s", cnf.Nats.Address[0].Host)
	log.Printf("Nats port: %s", cnf.Nats.Address[0].Port)
	log.Printf("Nats address: %s", cnf.Nats.Address[0].Address)
	log.Printf("Nats address: %s", cnf.GetNatsAddresses())
}

func TestSetup(t *testing.T) {
	path := "settings.json"

	set := NewApiSetup(path)

	log.Printf("%+v", set)
	log.Printf("Name: %s", set.Config.Name)
	log.Printf("Version: %s", set.Config.Version)
	log.Printf("Nats version: %s", set.Config.Nats.Version)
	log.Printf("Nats ReconnectedWait: %d", set.Config.Nats.ReconnectedWait)
	log.Printf("Nats host: %s", set.Config.Nats.Address[0].Host)
	log.Printf("Nats port: %s", set.Config.Nats.Address[0].Port)
	log.Printf("Nats address: %s", set.Config.Nats.Address[0].Address)
	log.Printf("Nats address: %s", set.Config.GetNatsAddresses())

	log.Println(set.Nats.ConnectedAddr())
}
