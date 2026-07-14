FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /cache-sidecar ./services/cache

FROM alpine:3.24
RUN apk add --no-cache ca-certificates
COPY --from=builder /cache-sidecar /cache-sidecar
COPY contracts/schemas/busca-contract.json /schemas/busca-contract.json
EXPOSE 50055
ENTRYPOINT ["/cache-sidecar"]
