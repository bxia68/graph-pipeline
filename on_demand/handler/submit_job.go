package handler

import (
    "encoding/json"
    "fmt"
    "strings"
    "io/ioutil"
    "net/http"
    "go.uber.org/zap"
    "github.com/gin-gonic/gin"
    "job_manager/pb"
)

type MapDescriptionData struct {
    LegendIds []uint64 `json:"legend_ids"`
}

type WeaviateData struct {
    ParagraphIds []string `json:"paragraph_ids"`
}

type SubmitJobRequest struct {
    MapDescriptionData MapDescriptionData `json:"map_description_data"`
    WeaviateData       WeaviateData       `json:"weaviate_data"`
}

type LegendResponse struct {
    LegendId uint64 `json:"legend_id"`
    Descrip  string `json:"descrip"`
}

func fetchDescriptions(legendId []uint64) (map[uint64]string, error) {
    ids := make([]string, len(legendId))
    for i, id := range legendId {
        ids[i] = fmt.Sprintf("%d", id)
    }
    url := fmt.Sprintf("https://dev2.macrostrat.org/api/pg/legend?select=legend_id,descrip&descrip=not.is.null&legend_id=in.(%s)", strings.Join(ids, ","))

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var legends []LegendResponse
    if err := json.Unmarshal(body, &legends); err != nil {
        return nil, err
    }

    descriptions := make(map[uint64]string)
    for _, legend := range legends {
        descriptions[legend.LegendId] = legend.Descrip
    }

    return descriptions, nil
}

func parseMapDescriptionJob(request *SubmitJobRequest) (*pb.GetJobResponse, error) {
    descriptionsMap, err := fetchDescriptions(request.MapDescriptionData.LegendIds)
    if err != nil {
        return nil, fmt.Errorf("Failed to fetch map descriptions")
    }

    descriptions := make([]*pb.MapDescription, len(request.MapDescriptionData.LegendIds))
    for i, id := range request.MapDescriptionData.LegendIds {
        descriptions[i] = &pb.MapDescription{LegendId: id, Text: descriptionsMap[id]}
    }

    return &pb.GetJobResponse{
        Type: pb.JobType_on_demand,
        JobData: &pb.GetJobResponse_MapDescriptionData{
            MapDescriptionData: &pb.MapDescriptionJob{
                Descriptions: descriptions,
            },
        },
    }, nil
}

func parseWeaviateJob(request *SubmitJobRequest)(*pb.GetJobResponse, error) {
    return &pb.GetJobResponse{
        Type: pb.JobType_on_demand,
        JobData: &pb.GetJobResponse_WeaviateData{
            WeaviateData: &pb.WeaviateJob{
                ParagraphIds: request.WeaviateData.ParagraphIds,
            },
        },
    }, nil
}


func HandleSubmitJob(logger *zap.Logger, r *gin.Engine, jobQueue chan *pb.GetJobResponse) {
	r.POST("/submit_job", func(c *gin.Context) {
        var request SubmitJobRequest
        if err := c.ShouldBindJSON(&request); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
		
        if len(request.MapDescriptionData.LegendIds) > 0 {
            job, err := parseMapDescriptionJob(&request)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse map data"})
                return
            }
            jobQueue <- job
        }

        if len(request.WeaviateData.ParagraphIds) > 0 {
            job, err := parseWeaviateJob(&request)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse weaviate data"})
                return
            }
            jobQueue <- job
        }

        c.JSON(http.StatusOK, gin.H{"status": "received"})
    })
}