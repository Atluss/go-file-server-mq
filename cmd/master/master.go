package main

import (
	"bytes"
	"fmt"
	"github.com/Atluss/FileServerWithMQ/pkg/v1/Transport"
	"github.com/Atluss/FileServerWithMQ/pkg/v1/config"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var tasks []Transport.Task
var taskMutex sync.Mutex
var oldestFinishedTaskPointer int

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

	tasks = make([]Transport.Task, 0, 20)
	taskMutex = sync.Mutex{}
	oldestFinishedTaskPointer = 0

	wg := sync.WaitGroup{}
	if _, err := set.Nats.Subscribe("Work.TaskToDo", func(m *nats.Msg) {

		myTaskPointer, ok := getNextTask()

		if ok {
			data, err := proto.Marshal(myTaskPointer)
			if err == nil {
				set.Nats.Publish(m.Reply, data)
			}
		}

	}); err != nil {
		log.Println(err)
	}

	if _, err := set.Nats.Subscribe("Work.TaskFinished", func(m *nats.Msg) {

		myTask := Transport.Task{}
		err := proto.Unmarshal(m.Data, &myTask)

		if err == nil {
			taskMutex.Lock()
			tasks[myTask.Id].State = 2
			tasks[myTask.Id].Finisheduuid = myTask.Finisheduuid
			taskMutex.Unlock()

		}

	}); err != nil {
		log.Println(err)
	}

	initTestTasks(set.Nats)

	wg.Add(1)
	wg.Wait()
}

func initTestTasks(nc *nats.Conn) {
	for i := 0; i < 20; i++ {

		newTask := Transport.Task{Uuid: uuid.NewV4().String(), State: 0}
		log.Printf("Gererate uid: %s", newTask.Uuid)
		fileServerAddressTransport := Transport.DiscoverableServiceTransport{}

		msg, err := nc.Request("Discovery.FileServer", nil, 1000*time.Millisecond)
		if err == nil && msg != nil {
			err := proto.Unmarshal(msg.Data, &fileServerAddressTransport)
			if err != nil {
				continue
			}
		}

		if err != nil {
			log.Printf("Something went wrong. (Discovery.FileServer) %s", err)
			continue
		}

		fileServerAddress := fileServerAddressTransport.Address
		data := make([]byte, 0, 1024)

		buf := bytes.NewBuffer(data)
		fmt.Fprint(buf, "get,my,data,my,get,get,have")

		r, err := http.Post(fileServerAddress+"/"+newTask.Uuid, "", buf)
		if err != nil || r.StatusCode != http.StatusOK {
			log.Printf("error send file: %s", err)
			continue
		}

		newTask.Id = int32(len(tasks))
		tasks = append(tasks, newTask)
	}
}

func getNextTask() (*Transport.Task, bool) {
	taskMutex.Lock()
	defer taskMutex.Unlock()

	for i := oldestFinishedTaskPointer; i < len(tasks); i++ {
		if i == oldestFinishedTaskPointer && tasks[i].State == 2 {
			oldestFinishedTaskPointer++
		} else {
			if tasks[i].State == 0 {
				tasks[i].State = 1
				go resetTaskIfNotFinished(i)
				return &tasks[i], true
			}
		}
	}
	return nil, false
}

func resetTaskIfNotFinished(i int) {
	time.Sleep(2 * time.Minute)
	taskMutex.Lock()

	if tasks[i].State != 2 {
		tasks[i].State = 0
	}
}
