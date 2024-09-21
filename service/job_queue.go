package service

import (
	"context"
	"job_manager/data_loaders"
	"job_manager/pb"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Generate jobs in groups and wait for group to finish before moving on to next group
func QueueJobs(ctx context.Context, logger *zap.Logger, jobQueue chan *pb.GetJobResponse, wg *sync.WaitGroup,
	groupSize, batchSize int, dataLoader data_loaders.DataLoader) {
	groupCount := 1
	jobCounter := uint64(0)
	for {
		groupStart := time.Now()
		for i := 0; i < groupSize; i++ {
			miniBatch, end := dataLoader.GetBatch(batchSize)
			miniBatch.Id = jobCounter
			miniBatch.Type = pb.JobType_batch
			jobCounter++
			jobQueue <- miniBatch

			wg.Add(1)
			if end {
				logger.Info("Queued all jobs.")
				wg.Wait()
				logger.Info("Finished all jobs.")
				return
			}
		}

		wg.Wait()
		logger.Info("Finished processing group",
			zap.Int("group", groupCount),
			zap.Float64("avg_seconds_per_paragraph", time.Since(groupStart).Seconds()/float64(groupSize*batchSize)),
		)
		groupCount++
	}
}
