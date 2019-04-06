package main

import (
	"bytes"
	"fmt"
	"github.com/Atluss/FileServerWithMQ/Transport"
	"github.com/Atluss/FileServerWithMQ/lib/config"
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

var Tasks []Transport.Task
var TaskMutex sync.Mutex
var oldestFinishedTaskPointer int

func main() {

	settingPath := "settings.json"
	set := config.NewApiSetup(settingPath)

	// do something if user close program (close DB, or wait running query)
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("Exit program...")
		os.Exit(1)
	}()

	Tasks = make([]Transport.Task, 0, 20)
	TaskMutex = sync.Mutex{}
	oldestFinishedTaskPointer = 0

	initTestTasks(set.Nats)

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
			TaskMutex.Lock()
			Tasks[myTask.Id].State = 2
			Tasks[myTask.Id].Finisheduuid = myTask.Finisheduuid
			TaskMutex.Unlock()

		}

	}); err != nil {
		log.Println(err)
	}

	wg.Add(1)
	wg.Wait()
}

func initTestTasks(nc *nats.Conn) {

	for i := 0; i < 20; i++ {

		newTask := Transport.Task{Uuid: uuid.NewV4().String(), State: 0}
		fileServerAddressTransport := Transport.DiscoverableServiceTransport{}

		msg, err := nc.Request("Discovery.FileServer", nil, 1000*time.Millisecond)
		if err == nil && msg != nil {
			err := proto.Unmarshal(msg.Data, &fileServerAddressTransport)
			if err != nil {
				continue
			}
		}

		if err != nil {
			continue
		}

		fileServerAddress := fileServerAddressTransport.Address
		data := make([]byte, 0, 1024)

		buf := bytes.NewBuffer(data)
		fmt.Fprint(buf, "get,my,data,my,get,get,have")

		r, err := http.Post(fileServerAddress+"/"+newTask.Uuid, "", buf)
		if err != nil || r.StatusCode != http.StatusOK {
			continue
		}

		newTask.Id = int32(len(Tasks))
		Tasks = append(Tasks, newTask)
	}
}

func getNextTask() (*Transport.Task, bool) {

	TaskMutex.Lock()
	defer TaskMutex.Unlock()

	for i := oldestFinishedTaskPointer; i < len(Tasks); i++ {
		if i == oldestFinishedTaskPointer && Tasks[i].State == 2 {
			oldestFinishedTaskPointer++
		} else {
			if Tasks[i].State == 0 {
				Tasks[i].State = 1
				go resetTaskIfNotFinished(i)
				return &Tasks[i], true
			}
		}
	}
	return nil, false

}

func resetTaskIfNotFinished(i int) {

	time.Sleep(2 * time.Minute)
	TaskMutex.Lock()

	if Tasks[i].State != 2 {
		Tasks[i].State = 0
	}
}
