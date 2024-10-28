FROM golang:latest

WORKDIR /app

RUN apt-get update && apt-get install -y protobuf-compiler
COPY go.mod go.sum ./

RUN go mod download
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN go get github.com/gin-gonic/gin

COPY pb pb
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pb/job_manager.proto

COPY coordinator coordinator
COPY on_demand on_demand
COPY server.go .

RUN go build -o server .

ENTRYPOINT ["./server"]