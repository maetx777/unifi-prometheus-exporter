FROM golang:alpine AS builder
WORKDIR /src
COPY . .
RUN apk add git
RUN cd cmd/exporter && go mod download
RUN go build -o bin/exporter ./cmd/exporter

FROM alpine:latest
RUN apk add ca-certificates
COPY --from=builder /src/bin/exporter /usr/bin/unifi-prometheus-exporter
CMD ["/usr/bin/unifi-prometheus-exporter"]