package data_loaders

import "job_manager/pb"

type DataLoader interface {
	Init(string)
	GetOffset() string
	GetBatch(int) (*pb.GetJobResponse, bool)
}