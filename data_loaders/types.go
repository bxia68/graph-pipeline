package data_loaders

import "job_manager/pb"

type DataLoader interface {
	Init()
	GetBatch(int) (*pb.GetJobResponse, bool)
}