FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0
ENV GO111MODULE=on

WORKDIR /go/src/github.com/segadora/solar-assistant-browser-automation

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o /usr/bin/solar-assistant-browser-automation .

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:3

RUN apk add curl
RUN apk add chromium

WORKDIR /
COPY --from=builder /usr/bin/solar-assistant-browser-automation /

EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=5s --start-period=10s --retries=3 CMD curl --fail http://localhost:1323/health || exit 1

CMD ["/solar-assistant-browser-automation"]
