package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

type MemoryAllocatedCustom struct {
	Instance        string
	MemoryAllocated map[string]float64
}
type memallocated struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name     string `json:"__name__"`
				Instance string `json:"instance"`
				Job      string `json:"job"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Memory allocated for the instance
func (m2 MemoryAllocatedCustom) GetMemAllocated() []*MemoryAllocatedCustom {

	var MA []*MemoryAllocatedCustom

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "http://"+os.Getenv("PROMETHEUS")+":"+os.Getenv("PROM_PORT")+"/api/v1/query?query=node_memory_MemTotal_bytes&g0.tab=1", nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	b := bodyText

	jsonfiles := make(map[string]interface{})
	var (
		a = memallocated{}
	)

	_ = json.Unmarshal(b, &a)
	jsonfiles["memallocated"] = a

	for _, item := range a.Data.Result {
		memallocated := &MemoryAllocatedCustom{}
		memallocated.MemoryAllocated = make(map[string]float64)
		memallocated.Instance = item.Metric.Instance
		memallocated.MemoryAllocated["MemAllocatedTimestamp"] = item.Value[0].(float64)

		memallocval := item.Value[1].(string)
		j, err := strconv.ParseFloat(memallocval, 64)
		if err != nil {
			log.Println(err)
		}
		memallocated.MemoryAllocated["MemAllocatedValue"] = j

		MA = append(MA, memallocated)
	}

	return MA
}
