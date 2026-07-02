# Context: repo root (docker build -f services/coupon-collector/Dockerfile .)
FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/coupon-collector ./services/coupon-collector

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/coupon-collector /coupon-collector
ENV GRPC_PORT=50061
EXPOSE 50061
ENTRYPOINT ["/coupon-collector"]
