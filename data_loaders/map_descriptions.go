package data_loaders

import (
	"encoding/json"
	"io"
	"job_manager/pb"
	"log"
	"net/http"
	"strconv"
)

type DescriptionsLoader struct {
	offset int
}

type Description struct {
	LegendID    uint64 `json:"legend_id"`
	Description string `json:"descrip"`
}

func (loader *DescriptionsLoader) Init() {
	loader.offset = 0
}

func (loader *DescriptionsLoader) GetBatch(batchSize int) (*pb.GetJobResponse, bool) {
	requestURL := "https://dev2.macrostrat.org/api/pg/legend"
	res, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Fatalf("error making http request: %s\n", err)
	}

	params := res.URL.Query()
	params.Set("select", "legend_id,descrip")
	params.Set("order", "legend_id.asc")
	params.Set("descrip", "not.is.null")
	params.Set("limit", strconv.Itoa(batchSize))
	params.Set("offset", strconv.Itoa(loader.offset))
	res.URL.RawQuery = params.Encode()

	resp, err := http.DefaultClient.Do(res)
	if err != nil {
		log.Fatalf("error sending http request: %s\n", err)
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("client: could not read response body: %s\n", err)
	}

	var descriptions []Description
	err = json.Unmarshal(resBody, &descriptions)
	if err != nil {
		log.Fatalf("error parsing JSON response: %s\n", err)
	}

	loader.offset += batchSize
	var parsed_list []*pb.MapDescription
	for _, d := range descriptions {
		parsed_list = append(parsed_list, &pb.MapDescription{
			LegendId: d.LegendID,
			Text:     d.Description,
		})
	}
	job := &pb.GetJobResponse{JobData: &pb.GetJobResponse_MapDescriptionData{MapDescriptionData: &pb.MapDescriptionJob{Descriptions: parsed_list}}}
	return job, len(parsed_list) < batchSize
}
