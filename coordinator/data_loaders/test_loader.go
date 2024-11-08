package data_loaders

import (
	"job_manager/pb"
	"strconv"
)

type TestLoader struct {
	data       []string
	batchIndex int
}

func (d *TestLoader) Init(offset string) {
	d.data = []string{
		"data1", "data2", "data3", "data4", "data5",
		"data6", "data7", "data8", "data9", "data10",
	}
	d.batchIndex, _ = strconv.Atoi(offset)
}

func (loader *TestLoader) GetOffset() string {
	return strconv.Itoa(loader.batchIndex)
}

func (d *TestLoader) GetBatch(batchSize int) (*pb.GetJobResponse, bool) {
	if d.batchIndex >= len(d.data) {
		return nil, true
	}

	endIndex := d.batchIndex + batchSize
	if endIndex > len(d.data) {
		endIndex = len(d.data)
	}

	batch := d.data[d.batchIndex:endIndex]
	d.batchIndex = endIndex

	job := &pb.GetJobResponse{JobData: &pb.GetJobResponse_TestData{TestData: &pb.TestJob{Paragraphs: batch}}}
	return job, d.batchIndex >= len(d.data)
}
