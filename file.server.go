package main

import (
	"fmt"
	"github.com/Atluss/FileServerWithMQ/Transport"
	"github.com/Atluss/FileServerWithMQ/lib/config"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/nats-io/go-nats"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	settingPath := "settings.json"

	set := config.NewApiSetup(settingPath)

	log.Printf("Name: %s", set.Config.Name)
	log.Printf("Version: %s", set.Config.Version)
	log.Printf("Nats version: %s", set.Config.Nats.Version)
	log.Printf("Nats ReconnectedWait: %d", set.Config.Nats.ReconnectedWait)
	log.Printf("Nats host: %s", set.Config.Nats.Address[0].Host)
	log.Printf("Nats port: %s", set.Config.Nats.Address[0].Port)
	log.Printf("Nats address: %s", set.Config.Nats.Address[0].Address)
	log.Printf("Nats address(multi): %s", set.Config.GetNatsAddresses())

	// do something if user close program (close DB, or wait running query)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Exit program...")
		os.Exit(1)
	}()

	set.Route.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		file, err := os.Open("tmp/" + vars["name"])
		defer file.Close()

		if err != nil {
			w.WriteHeader(404)
		}

		if file != nil {
			_, err := io.Copy(w, file)
			if err != nil {
				w.WriteHeader(500)
			}
		}

	}).Methods("GET")

	set.Route.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		file, err := os.Create("tmp/" + vars["name"])
		defer file.Close()

		if err != nil {
			w.WriteHeader(500)
		}

		if file != nil {
			_, err := io.Copy(file, r.Body)
			if err != nil {
				w.WriteHeader(500)
			} else {
				log.Printf("File: %s uploaded", vars["name"])
			}
		}

	}).Methods("POST")

	RunServiceDiscoverable(set.Nats, set.Config.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", set.Config.Port), set.Route); err != nil {
		log.Println(err)
	}
}

func RunServiceDiscoverable(nc *nats.Conn, port string) {
	if _, err := nc.Subscribe("Discovery.FileServer", func(m *nats.Msg) {

		serviceAddressTransport := Transport.DiscoverableServiceTransport{Address: fmt.Sprintf("http://localhost:%s", port)}
		data, err := proto.Marshal(&serviceAddressTransport)
		if err == nil {
			nc.Publish(m.Reply, data)
		}
	}); err != nil {
		log.Println(err)
	}
}
