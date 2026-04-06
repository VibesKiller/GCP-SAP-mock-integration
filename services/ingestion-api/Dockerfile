FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/ingestion-api ./services/ingestion-api

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/ingestion-api /app/ingestion-api
EXPOSE 8081
ENTRYPOINT ["/app/ingestion-api"]
