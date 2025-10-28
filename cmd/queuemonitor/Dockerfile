FROM golang:1.25.3-alpine3.22 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o bin/queuemonitor ./cmd/queuemonitor

FROM alpine:3.22 AS final

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /app/bin/queuemonitor .

CMD ["./queuemonitor"]