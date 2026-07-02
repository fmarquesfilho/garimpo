# Context: repo root (docker build -f services/collector-amazon/Dockerfile .)
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/collector-amazon ./services/collector-amazon

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/collector-amazon /collector-amazon
ENV PORT=50055
EXPOSE 50055
ENTRYPOINT ["/collector-amazon"]
