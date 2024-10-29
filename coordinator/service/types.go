package service

import (
	"job_manager/pb"
	"time"
)

type JobType string

const (
	Wait        JobType = "wait"
	BatchJob    JobType = "batch"
	OnDemandJob JobType = "on_demand"
)

type UpdateType string

const (
	ManageHealth UpdateType = "manage_health"
	UpdateHealth UpdateType = "update_health"
	RequestJob   UpdateType = "request"
	FinishJob    UpdateType = "finish"
)

type StateUpdate struct {
	UpdateType UpdateType
	JobId      uint64
	JobData    Job
}

type Job struct {
	*pb.GetJobResponse
	HealthTimeout int
	StartTime     time.Time
	HealthTime    time.Time
}
