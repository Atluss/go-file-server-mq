package main

import (
	"bytes"
	"fmt"
	"github.com/Atluss/FileServerWithMQ/pkg/v1/Transport"
	"github.com/Atluss/FileServerWithMQ/pkg/v1/config"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	settingPath := "cmd/settings.json"

	set := config.NewApiSetup(settingPath)

	// do something if user close program (close DB, or wait running query)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Exit program...")
		os.Exit(1)
	}()

	for i := 0; i < 8; i++ {
		go doWork(set.Nats)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func doWork(nc *nats.Conn) {
	for {
		// We ask for a Task with a 1 second Timeout
		msg, err := nc.Request("Work.TaskToDo", nil, 3*time.Second)
		if err != nil {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s (Work.TaskToDo)", err)
			continue
		}

		// We unmarshal the Task
		curTask := Transport.Task{}
		err = proto.Unmarshal(msg.Data, &curTask)
		if err != nil {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s", err)
			continue
		}

		// We get the FileServer address
		msg, err = nc.Request("Discovery.FileServer", nil, 1000*time.Millisecond)
		if err != nil {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s", err)
			continue
		}

		fileServerAddressTransport := Transport.DiscoverableServiceTransport{}
		err = proto.Unmarshal(msg.Data, &fileServerAddressTransport)
		if err != nil {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s", err)
			continue
		}

		// We get the file
		fileServerAddress := fileServerAddressTransport.Address
		r, err := http.Get(fileServerAddress + "/" + curTask.Uuid)
		if err != nil {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s", err)
			continue
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s", err)
			continue
		}

		// We split and count the words
		words := strings.Split(string(data), ",")
		sort.Strings(words)
		wordCounts := make(map[string]int)
		for i := 0; i < len(words); i++ {
			wordCounts[words[i]] = wordCounts[words[i]] + 1
		}

		resultData := make([]byte, 0, 1024)
		buf := bytes.NewBuffer(resultData)
		// We print the results to a buffer
		for key, value := range wordCounts {
			fmt.Fprintln(buf, key, ":", value)
		}

		// We generate a new UUID for the finished file
		curTask.Finisheduuid = uuid.NewV4().String()
		r, err = http.Post(fileServerAddress+"/"+curTask.Finisheduuid, "", buf)
		if err != nil || r.StatusCode != http.StatusOK {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s:%d", err, r.StatusCode)
			continue
		}

		// We marshal the current Task into a protobuffer
		data, err = proto.Marshal(&curTask)
		if err != nil {
			log.Printf("Something went wrong. Waiting 2 seconds before retrying: %s", err)
			continue
		}

		log.Println("all okey")

		// We notify the Master about finishing the Task
		nc.Publish("Work.TaskFinished", data)
	}
}
