package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
)

func main() {
	hubConnStr := os.Getenv("AZ_HUB_CONN_STR")
	log.Printf("Using connection string from environment variable: %s", hubConnStr)
	hub, err := eventhub.NewHubFromConnectionString(hubConnStr)
	if err != nil {
		log.Panic(err)
		return
	}

	for {
		loadavg := getSysLoad()
		log.Printf("Sending system load data point: %v", loadavg)
		err := sendDataPointToHub(hub, loadavg)
		log.Printf("Send error: %v", err)
		time.Sleep(10 * time.Second)
	}
}

func sendDataPointToHub(hub *eventhub.Hub, loadavg float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pid := os.Getpid()
	hostName, _ := os.Hostname()

	dataPoint := struct {
		PID         int     `json:"PID"`
		HostName    string  `json:"HostName"`
		UnixSec     int64   `json:"UnixSec"`
		LoadAverage float64 `json:"LoadAverage"`
	}{
		PID:         pid,
		HostName:    hostName,
		UnixSec:     time.Now().Unix(),
		LoadAverage: loadavg,
	}
	jsonDoc, err := json.Marshal(dataPoint)
	if err != nil {
		log.Panic(err)
	}
	return hub.Send(ctx, eventhub.NewEventFromString(string(jsonDoc)))
}

func getSysLoad() float64 {
	loadavgContnet, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		log.Panic(err)
	}
	fields := strings.Split(string(loadavgContnet), " ")
	if len(fields) < 1 {
		log.Panic("failed to interpret the content of loadavg")
	}
	val, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		log.Panic(err)
	}
	return val
}
