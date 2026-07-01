# Context: repo root (docker build -f services/publisher/Dockerfile .)
FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/publisher ./services/publisher

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/publisher /publisher
ENV PORT=50052
EXPOSE 50052
ENTRYPOINT ["/publisher"]
