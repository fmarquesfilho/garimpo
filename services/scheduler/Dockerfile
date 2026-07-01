# Context: repo root (docker build -f services/scheduler/Dockerfile .)
FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/scheduler ./services/scheduler

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/scheduler /scheduler
ENV PORT=50054
EXPOSE 50054
ENTRYPOINT ["/scheduler"]
