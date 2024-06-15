FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o ./bin/scribe ./cmd/scribe

CMD ["./bin/scribe"]
