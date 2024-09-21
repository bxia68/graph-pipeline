package main

import (
	"context"
	"flag"
	"fmt"
	"job_manager/data_loaders"
	"job_manager/service"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"job_manager/pb"

	"go.uber.org/zap"
)

type server struct {
	pb.UnimplementedJobManagerServer
	stateChan  chan service.StateUpdate
	jobQueue   chan *pb.GetJobResponse
	pipelineID string
	runID      string
	logger     *zap.Logger
}

func (s *server) GetJob(ctx context.Context, args *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	var job *pb.GetJobResponse
	select {
	case job = <-s.jobQueue:
	case <-time.After(1 * time.Second):
		return &pb.GetJobResponse{Type: pb.JobType_wait}, nil
	}

	currentTime := time.Now()
	s.stateChan <- service.StateUpdate{
		UpdateType: service.RequestJob,
		JobId:      job.Id,
		JobData: service.Job{
			GetJobResponse: job,
			HealthTime:     currentTime,
			StartTime:      currentTime,
		},
	}

	return job, nil
}

func (s *server) FinishJob(ctx context.Context, args *pb.FinishJobRequest) (*pb.FinishJobResponse, error) {
	s.stateChan <- service.StateUpdate{
		UpdateType: service.FinishJob,
		JobId:      args.GetId(),
	}
	return &pb.FinishJobResponse{}, nil
}

func (s *server) UpdateHealth(ctx context.Context, args *pb.UpdateHealthRequest) (*pb.UpdateHealthResponse, error) {
	s.stateChan <- service.StateUpdate{
		UpdateType: service.UpdateHealth,
		JobId:      args.GetId(),
	}
	return &pb.UpdateHealthResponse{}, nil
}

func (s *server) GetMetadata(ctx context.Context, args *pb.GetMetadataRequest) (*pb.GetMetadataResponse, error) {
	return &pb.GetMetadataResponse{
		RunId:      s.runID,
		PipelineId: s.pipelineID,
	}, nil
}

func InitLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize zap logger: %v", err))
	}
	return logger
}

func main() {
	ctx := context.Background()

	logger := InitLogger()
	defer logger.Sync()

	groupSize, _ := strconv.Atoi(os.Getenv("GROUP_SIZE"))
	miniBatchSize, _ := strconv.Atoi(os.Getenv("MINI_BATCH_SIZE"))
	interval, _ := strconv.Atoi(os.Getenv("HEALTH_INTERVAL"))
	timeout, _ := strconv.Atoi(os.Getenv("HEALTH_TIMEOUT"))
	if interval == 0 || timeout == 0 {
		logger.Fatal("Invalid health interval or timeout")
	}
	healthInterval := time.Duration(interval) * time.Second
	healthTimeout := time.Duration(timeout) * time.Second

	pipelineID := os.Getenv("PIPELINE_ID")
	runID := os.Getenv("RUN_ID")
	runID += time.Now().Format("_2006-01-02_15:04:05")
	logger.Info(runID)
	loader_flag := flag.String("loader", "test", "Specify the data loader for the manager. Options are \"weaviate\", \"map_descrip\", \"test\".")
	flag.Parse()

	var loader data_loaders.DataLoader
	switch *loader_flag {
	// case "weaviate":
	// 	loader = &data_loaders.WeaviateLoader{}
	case "map_descrip":
		loader = &data_loaders.DescriptionsLoader{}
	case "test":
		loader = &data_loaders.TestLoader{}
	default:
		logger.Fatal("Invalid dataloader")
	}
	loader.Init()

	stateChan := make(chan service.StateUpdate)
	jobQueue := make(chan *pb.GetJobResponse, miniBatchSize)
	var wg sync.WaitGroup
	go service.ManageJobStates(ctx, logger, stateChan, jobQueue, &wg, healthInterval, healthTimeout)
	go service.QueueJobs(ctx, logger, jobQueue, &wg, groupSize, miniBatchSize, loader)

	s := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	pb.RegisterJobManagerServer(s, &server{
		pipelineID: pipelineID,
		runID:      runID,
		logger:     logger,
		stateChan:  stateChan,
		jobQueue:   jobQueue,
	})
	logger.Info("server listening", zap.String("address", lis.Addr().String()))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}
