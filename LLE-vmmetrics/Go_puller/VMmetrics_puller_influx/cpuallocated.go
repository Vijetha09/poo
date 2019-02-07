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

type CpuAllocatedCustom struct {
	Instance     string
	CpuAllocated map[string]float64
}

type cpuallocated struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Instance string `json:"instance"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// gives CPU allocated for the instance
func (c1 CpuAllocatedCustom) GetCpuAllocated() []*CpuAllocatedCustom {

	var CB []*CpuAllocatedCustom

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "http://"+os.Getenv("PROMETHEUS")+":"+os.Getenv("PROM_PORT")+"/api/v1/query?query=ceil(sum%20by%20(idle%2C%20instance)%20(irate(node_cpu_seconds_total%7Bjob%3D%22prometheus%22%7D%5B1m%5D)))&g0.tab=1", nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	b := bodyText

	jsonfiles := make(map[string]interface{})
	var (
		a = cpuallocated{}
	)

	_ = json.Unmarshal(b, &a)
	jsonfiles["cpualloc"] = a

	for _, item := range a.Data.Result {
		cpuallocted := &CpuAllocatedCustom{}
		cpuallocted.CpuAllocated = make(map[string]float64)
		cpuallocted.Instance = item.Metric.Instance
		cpuallocted.CpuAllocated["CpuAllocatedTimestamp"] = item.Value[0].(float64)

		cpuaval := item.Value[1].(string)
		j, err := strconv.ParseFloat(cpuaval, 64)
		if err != nil {
			log.Println(err)
		}
		cpuallocted.CpuAllocated["CpuAllocatedValue"] = j

		CB = append(CB, cpuallocted)
	}

	return CB
}
