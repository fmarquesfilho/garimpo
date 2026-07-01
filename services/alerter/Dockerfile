# Context: repo root (docker build -f services/alerter/Dockerfile .)
FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/alerter ./services/alerter

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/alerter /alerter
ENV PORT=50053
EXPOSE 50053
ENTRYPOINT ["/alerter"]
