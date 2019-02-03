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

type CpuUsageCustom struct {
	Instance string
	CpuUsed  map[string]float64
}
type cpuusage struct {
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

// CPU usage of the instance
func (c2 CpuUsageCustom) GetCpuUsage() []*CpuUsageCustom {

	var CU []*CpuUsageCustom

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "http://"+os.Getenv("PROMETHEUS")+":"+os.Getenv("PROM_PORT")+"/api/v1/query?query=100-(avg%20by%20(instance)%20(irate(node_cpu_seconds_total%7Bjob%3D%22prometheus%22%2Cmode%3D%22idle%22%7D%5B1m%5D))%20*%20100)&g0.tab=1", nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	b := bodyText

	jsonfiles := make(map[string]interface{})
	var (
		a = cpuusage{}
	)

	_ = json.Unmarshal(b, &a)
	jsonfiles["cpuusage"] = a

	for _, item := range a.Data.Result {
		cpuusage := &CpuUsageCustom{}
		cpuusage.CpuUsed = make(map[string]float64)
		cpuusage.Instance = item.Metric.Instance
		cpuusage.CpuUsed["CpuUsedTimestamp"] = item.Value[0].(float64)

		cpuusageval := item.Value[1].(string)
		j, err := strconv.ParseFloat(cpuusageval, 64)
		if err != nil {
			log.Println(err)
		}
		cpuusage.CpuUsed["CpuUsedValue"] = j

		CU = append(CU, cpuusage)
	}

	return CU
}
