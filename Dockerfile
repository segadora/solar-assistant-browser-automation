FROM golang:1.23-alpine AS build

WORKDIR /app

COPY *.go ./
COPY go.mod go.sum ./

RUN go get -d -v ./...

RUN CGO_ENABLED=0 GOOS=linux go build -o solar-assistant *.go

FROM alpine:3 AS executable

COPY --from=build  /app/solar-assistant /

RUN apk add curl
RUN apk add chromium

HEALTHCHECK --interval=5s --timeout=5s --start-period=10s --retries=3 CMD curl --fail http://localhost:1323/health || exit 1

ENTRYPOINT ["/solar-assistant"]
