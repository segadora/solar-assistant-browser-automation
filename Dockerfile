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
USER nonroot:nonroot

CMD ["/solar-assistant-browser-automation"]
