FROM golang:alpine AS builder
WORKDIR /build
ADD go.mod .
COPY . .
RUN go build -o exporter exporter.go
FROM alpine
WORKDIR /
COPY --from=builder /build/exporter /exporter
CMD ["./exporter"]