package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"job_manager/on_demand/handler"
	"job_manager/pb"
)

func AddEndpoints(logger *zap.Logger, r *gin.Engine, jobQueue chan *pb.GetJobResponse) {
	handler.HandleSubmitJob(logger, r, jobQueue)
}

func StartOnDemandServer(logger *zap.Logger, jobQueue chan *pb.GetJobResponse) {
	r := gin.Default()

	AddEndpoints(logger, r, jobQueue)

	logger.Info("on demand server listening")
	r.Run(":8080")
}
