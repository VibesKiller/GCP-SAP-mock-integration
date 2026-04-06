FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/sap-mock-api ./services/sap-mock-api

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/sap-mock-api /app/sap-mock-api
EXPOSE 8080
ENTRYPOINT ["/app/sap-mock-api"]
