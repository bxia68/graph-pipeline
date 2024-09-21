package service

import (
	"context"
	"job_manager/pb"
	"sync"
	"time"

	"go.uber.org/zap"
)

func ManageJobStates(ctx context.Context, logger *zap.Logger, stateChan chan StateUpdate, jobQueue chan *pb.GetJobResponse,
	wg *sync.WaitGroup, healthInterval, healthTimeout time.Duration) {
	jobMap := make(map[uint64]*Job)

	go RunRoutineHealthCheck(ctx, stateChan, healthInterval, logger)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-stateChan:
			switch update.UpdateType {
			case UpdateHealth:
				jobMap[update.JobId].HealthTime = time.Now()
			case ManageHealth:
				ManageJobHealth(jobMap, jobQueue, healthTimeout, logger)
			case RequestJob:
				jobMap[update.JobId] = &update.JobData
			case FinishJob:
				delete(jobMap, update.JobId)
				wg.Done()
			}
		}
	}
}

// Scan through all running jobs and requeue job if it hasn't received a health check in healthTimeout seconds
func ManageJobHealth(jobMap map[uint64]*Job, jobQueue chan *pb.GetJobResponse, healthTimeout time.Duration, logger *zap.Logger) {
	currentTime := time.Now()
	for _, job := range jobMap {
		if job.HealthTime.Before(currentTime.Add(-healthTimeout)) {
			logger.Warn("Job has timed out",
				zap.Uint64("job_id", job.Id),
			)
			jobQueue <- job.GetJobResponse
			delete(jobMap, job.Id)
		}
	}
}

func RunRoutineHealthCheck(ctx context.Context, stateChan chan StateUpdate, healthInterval time.Duration, logger *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(healthInterval):
			stateChan <- StateUpdate{UpdateType: ManageHealth}
		}
	}
}
