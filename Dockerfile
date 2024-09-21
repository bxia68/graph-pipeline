FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY data_loaders data_loaders
COPY server.go .

RUN go build -o server .

ENTRYPOINT ["./server"]