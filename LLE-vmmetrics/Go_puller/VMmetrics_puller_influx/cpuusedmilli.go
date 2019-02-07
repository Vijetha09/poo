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

type CpuUsageMilliCustom struct {
	Instance     string
	CpuUsedMilli map[string]float64
}
type cpuusagemilli struct {
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
// CPU usage in milli cores 
func (c2 CpuUsageMilliCustom) GetCpuUsageMilli() []*CpuUsageMilliCustom {

	var CM []*CpuUsageMilliCustom
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "http://"+os.Getenv("PROMETHEUS")+":"+os.Getenv("PROM_PORT")+"/api/v1/query?query=(100-(avg%20by%20(instance)%20(irate(node_cpu_seconds_total%7Bjob%3D%22prometheus%22%2Cmode%3D%22idle%22%7D%5B1m%5D))%20*%20100))*(ceil(sum%20by%20(idle%2C%20instance)%20(irate(node_cpu_seconds_total%7Bjob%3D%22prometheus%22%7D%5B1m%5D))))*10&g0.tab=1", nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	b := bodyText

	jsonfiles := make(map[string]interface{})
	var (
		a = cpuusagemilli{}
	)

	_ = json.Unmarshal(b, &a)
	jsonfiles["cpuusagemilli"] = a

	for _, item := range a.Data.Result {
		cpuusagemilli := &CpuUsageMilliCustom{}
		cpuusagemilli.CpuUsedMilli = make(map[string]float64)
		cpuusagemilli.Instance = item.Metric.Instance
		cpuusagemilli.CpuUsedMilli["CpuUsedMilliTimestamp"] = item.Value[0].(float64)

		cpuusagemillival := item.Value[1].(string)
		j, err := strconv.ParseFloat(cpuusagemillival, 64)
		if err != nil {
			log.Println(err)
		}
		cpuusagemilli.CpuUsedMilli["CpuUsedMilliValue"] = j

		CM = append(CM, cpuusagemilli)
	}

	return CM
}
