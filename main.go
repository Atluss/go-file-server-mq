package main

import (
	"fmt"
	"github.com/Atluss/FileServerWithMQ/Transport"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"os"
)

func main() {

}

func RunServiceDiscoverable() {

	nc, err := nats.Connect(os.Args[1])

	if err != nil {

		fmt.Println("Can't connect to NATS. Service is not discoverable.")

	}

	nc.Subscribe("Discovery.FileServer", func(m *nats.Msg) {

		serviceAddressTransport := Transport.DiscoverableServiceTransport{Address: "http://localhost:3000"}

		data, err := proto.Marshal(&serviceAddressTransport)

		if err == nil {

			nc.Publish(m.Reply, data)

		}

	})

}
