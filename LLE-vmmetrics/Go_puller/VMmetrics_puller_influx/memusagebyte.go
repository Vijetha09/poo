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

type MemoryUsedBytesCustom struct {
	Instance        string
	MemoryUsedBytes map[string]float64
}
type memusedbyte struct {
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

func (m3 MemoryUsedBytesCustom) GetMemUsedBytes() []*MemoryUsedBytesCustom {

	var MU []*MemoryUsedBytesCustom

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", "http://"+os.Getenv("PROMETHEUS")+":"+os.Getenv("PROM_PORT")+"/api/v1/query?query=(node_memory_MemTotal_bytes-(node_memory_Buffers_bytes%2Bnode_memory_Cached_bytes%2Bnode_memory_MemFree_bytes%2Bnode_memory_Slab_bytes))*on(instance)group_left(nodename)(node_uname_info)&g0.tab=1", nil)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	b := bodyText

	jsonfiles := make(map[string]interface{})
	var (
		a = memusedbyte{}
	)

	_ = json.Unmarshal(b, &a)
	jsonfiles["memusage"] = a

	for _, item := range a.Data.Result {
		memusedbyte := &MemoryUsedBytesCustom{}
		memusedbyte.MemoryUsedBytes = make(map[string]float64)
		memusedbyte.Instance = item.Metric.Instance
		memusedbyte.MemoryUsedBytes["MemUsedByteTimestamp"] = item.Value[0].(float64)

		memusedval := item.Value[1].(string)
		j, err := strconv.ParseFloat(memusedval, 64)
		if err != nil {
			log.Println(err)
		}
		memusedbyte.MemoryUsedBytes["MemUsedByteValue"] = j

		MU = append(MU, memusedbyte)
	}

	return MU
}
