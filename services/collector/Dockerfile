# Context: repo root (docker build -f services/collector/Dockerfile .)
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/collector ./services/collector

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/collector /collector
ENV PORT=50051
EXPOSE 50051
ENTRYPOINT ["/collector"]
