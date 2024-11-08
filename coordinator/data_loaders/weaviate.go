package data_loaders

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/auth"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
	"log"
	"os"
	"job_manager/pb"
)

type ParagraphData struct {
	Additional struct {
		ID string `json:"id"`
	} `json:"_additional"`
	TopicList []string `json:"topic_list"`
}

type GraphQLData struct {
	Paragraphs []ParagraphData `json:"Paragraph"`
}

type WeaviateLoader struct {
	weaviate_client *weaviate.Client
	cursor          string
}

func (loader *WeaviateLoader) Init(offset string) {
	host := fmt.Sprintf("%v:%v", os.Getenv("WEAVIATE_HOST"), os.Getenv("WEAVIATE_PORT"))
	cfg := weaviate.Config{
		Host:       host,
		Scheme:     "http",
		AuthConfig: auth.ApiKey{Value: os.Getenv("WEAVIATE_API_KEY")},
		Headers:    nil,
	}

	var err error
	loader.weaviate_client, err = weaviate.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	loader.cursor = offset
}

func (loader *WeaviateLoader) GetOffset() string {
	return loader.cursor
}

func (loader *WeaviateLoader) GetBatch(batchSize int) (*pb.GetJobResponse, bool) {
	var batch []string
	for {
		response, err := getBatchWithCursor(loader.weaviate_client, batchSize, loader.cursor)
		if err != nil {
			log.Fatalf("Error fetching data or GraphQL errors: %v", err)
		}

		// parse response
		var data GraphQLData
		responseData, err := json.Marshal(response.Data["Get"])
		if err != nil {
			log.Fatalf("Error converting response data to byte slice: %v", err)
		}
		if err := json.Unmarshal(responseData, &data); err != nil {
			log.Fatalf("Error parsing GraphQL data: %v", err)
		}

		// filter to only geoarchive topics
		for _, paragraph := range data.Paragraphs {
			for _, topic := range paragraph.TopicList {
				if topic == "geoarchive" {
					batch = append(batch, paragraph.Additional.ID)
					continue
				}
			}
		}

		// update cursor
		loader.cursor = data.Paragraphs[len(data.Paragraphs)-1].Additional.ID

		// check if batch is full or if there are no more paragraphs left in weaviate
		if len(data.Paragraphs) < batchSize || len(batch) >= batchSize {
			response := &pb.GetJobResponse{JobData: &pb.GetJobResponse_WeaviateData{WeaviateData: &pb.WeaviateJob{ParagraphIds: batch}}}
			return response, len(data.Paragraphs) < batchSize
		}
	}
}

func getBatchWithCursor(client *weaviate.Client, batchSize int, cursor string) (*models.GraphQLResponse, error) {
	get := client.GraphQL().Get().
		WithClassName("Paragraph").
		WithFields(graphql.Field{Name: "_additional { id }"}, graphql.Field{Name: "topic_list"}).
		WithLimit(batchSize)

	if cursor != "" {
		return get.WithAfter(cursor).Do(context.Background())
	}
	return get.Do(context.Background())
}
