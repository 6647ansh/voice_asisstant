FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./
RUN go env -w GO111MODULE=on

COPY . .

RUN go build -o orchestrator main.go

EXPOSE 8080
CMD ["/app/orchestrator"]
