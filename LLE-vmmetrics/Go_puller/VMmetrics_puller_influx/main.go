package main

import (
	"log"
	"os"
	"os/signal"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/robfig/cron"
)

var fields map[string]interface{}

var Influx_client client.Client
var err error

func main() {

	// Create a new HTTPClient
	Influx_client, err = client.NewHTTPClient(client.HTTPConfig{
		// ENV variable values will be set in docker image
		// Addr: "http://localhost:8086",
		Addr: "http://" + os.Getenv("INFLUX") + ":" + os.Getenv("INFLUX_PORT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer Influx_client.Close()

	c1 := cron.New()
	//Adds functions to running cron and for evry one minute it will be executed
	c1.AddFunc("1 * * * * *", VmInfo)

	go c1.Start()
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig

}
