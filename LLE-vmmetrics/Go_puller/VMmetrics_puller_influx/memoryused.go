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

type MemoryUsedCustom struct {
	Instance   string
	NodeName   string
	MemoryUsed map[string]float64
}

type memused struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Instance string `json:"instance"`
				Job      string `json:"job"`
				Nodename string `json:"nodename"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (m1 MemoryUsedCustom) GetmemoryUsed() []*MemoryUsedCustom {

	var MU []*MemoryUsedCustom

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "http://"+os.Getenv("PROMETHEUS")+":"+os.Getenv("PROM_PORT")+"/api/v1/query?query=(((node_memory_MemTotal_bytes-(node_memory_Buffers_bytes%2Bnode_memory_Cached_bytes%2Bnode_memory_MemFree_bytes%2Bnode_memory_Slab_bytes))%2Fnode_memory_MemTotal_bytes)*100)*on(instance)group_left(nodename)(node_uname_info)&g0.tab=1", nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	b := bodyText

	jsonfiles := make(map[string]interface{})
	var (
		a = memused{}
	)

	_ = json.Unmarshal(b, &a)
	jsonfiles["memused"] = a

	for _, item := range a.Data.Result {
		memused := &MemoryUsedCustom{}
		memused.MemoryUsed = make(map[string]float64)
		memused.Instance = item.Metric.Instance
		memused.NodeName = item.Metric.Nodename
		memused.MemoryUsed["MemoryUsedTimestamp"] = item.Value[0].(float64)

		memuseval := item.Value[1].(string)
		j, err := strconv.ParseFloat(memuseval, 64)
		if err != nil {
			log.Println(err)
		}
		memused.MemoryUsed["MemoryUsedValue"] = j

		MU = append(MU, memused)
	}

	return MU
}
