FROM golang:1.24.4-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o bin/queuemonitor ./cmd/queuemonitor

FROM alpine:3.22 AS final

WORKDIR /app

COPY --from=build /app/bin/queuemonitor .

CMD ["./queuemonitor"]