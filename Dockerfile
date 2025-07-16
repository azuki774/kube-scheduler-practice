# Builder Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN apk add --no-cache make bash
RUN make bin-docker

# Runtime Stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/kube-scheduler-practice /app/kube-scheduler-practice

ENTRYPOINT ["/app/kube-scheduler-practice", "start"]
