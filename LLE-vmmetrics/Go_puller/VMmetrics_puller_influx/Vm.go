package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
)

type VmMetrics struct {
	MetricList []*Final
}

type Final struct {
	Instance        string
	NodeName        string
	CpuAllocated    map[string]float64
	CpuUsed         map[string]float64
	CpuUsedMilli    map[string]float64
	MemoryAllocated map[string]float64
	MemoryUsed      map[string]float64
	MemoryUsedBytes map[string]float64
}

// cpu ,memory information for VM instances
func VmInfo() {

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{

		Database: os.Getenv("DATABASE"),
	})
	if err != nil {
		log.Fatal(err)
	}

	cpuAllocatedDetails := CpuAllocatedCustom{}
	cpuAllocatedList := []*CpuAllocatedCustom{}

	cpuUsageDetails := CpuUsageCustom{}
	cpuUsageList := []*CpuUsageCustom{}

	cpuUsageMilliDetails := CpuUsageMilliCustom{}
	cpuUsageMilliList := []*CpuUsageMilliCustom{}

	memoryAllocatedDetails := MemoryAllocatedCustom{}
	memoryAllocatedList := []*MemoryAllocatedCustom{}

	memoryUsedDetails := MemoryUsedCustom{}
	memoryUsedList := []*MemoryUsedCustom{}

	memoryUsedBytesDetails := MemoryUsedBytesCustom{}
	memoryUsedBytesList := []*MemoryUsedBytesCustom{}

	var FL []*Final
	cpuAllocatedList = cpuAllocatedDetails.GetCpuAllocated()
	cpuUsageList = cpuUsageDetails.GetCpuUsage()
	cpuUsageMilliList = cpuUsageMilliDetails.GetCpuUsageMilli()

	memoryAllocatedList = memoryAllocatedDetails.GetMemAllocated()
	memoryUsedList = memoryUsedDetails.GetmemoryUsed()
	memoryUsedBytesList = memoryUsedBytesDetails.GetMemUsedBytes()

	for _, memUsageData := range memoryUsedList {

		FinalList := &Final{}
		FinalList.Instance = memUsageData.Instance
		FinalList.NodeName = memUsageData.NodeName

		FinalList.MemoryUsed = memUsageData.MemoryUsed
		FL = append(FL, FinalList)

	}

	for _, finalListDat := range FL {
		for _, cpuUsageData := range cpuUsageList {
			if finalListDat.Instance == cpuUsageData.Instance {
				finalListDat.CpuUsed = cpuUsageData.CpuUsed
			}
		}
	}
	for _, finalListDat := range FL {
		for _, cpuUsageMilliData := range cpuUsageMilliList {
			if finalListDat.Instance == cpuUsageMilliData.Instance {
				finalListDat.CpuUsedMilli = cpuUsageMilliData.CpuUsedMilli
			}
		}
	}

	for _, finalListDat := range FL {
		for _, cpuAllocData := range cpuAllocatedList {
			if finalListDat.Instance == cpuAllocData.Instance {
				finalListDat.CpuAllocated = cpuAllocData.CpuAllocated
			}
		}
	}

	for _, finalListDat := range FL {
		for _, memAllocData := range memoryAllocatedList {
			if finalListDat.Instance == memAllocData.Instance {
				finalListDat.MemoryAllocated = memAllocData.MemoryAllocated
			}
		}
	}

	for _, finalListDat := range FL {
		for _, memAllocData := range memoryUsedBytesList {
			if finalListDat.Instance == memAllocData.Instance {
				finalListDat.MemoryUsedBytes = memAllocData.MemoryUsedBytes
			}
		}
	}

	VmMetrics := VmMetrics{}
	VmMetrics.MetricList = FL

	jsonData, err := json.Marshal(VmMetrics)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(jsonData))

	// Create a point and add to batch

	tags := map[string]string{"VM": "vm-vinfo"}
	fields = make(map[string]interface{})
	fields["vminfo"] = string(jsonData)

	pt, err := client.NewPoint("vminfo", tags, fields, time.Now())

	if err != nil {
		log.Fatal(err)
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := Influx_client.Write(bp); err != nil {
		log.Fatal(err)
	}

	// Close client resources
	if err := Influx_client.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(".....data written into InfluxDb........\n")

}
