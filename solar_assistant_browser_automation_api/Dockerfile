ARG BUILD_ARCH
ARG BUILD_FROM

FROM --platform=${BUILD_ARCH:-linux/amd64} golang:1.23 AS builder

WORKDIR /go/src/github.com/segadora/solar-assistant-browser-automation

RUN echo v1.0.0
RUN git clone https://github.com/segadora/solar-assistant-browser-automation.git .

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=${BUILD_ARCH} go build -a -installsuffix cgo -o /usr/bin/solar-assistant-browser-automation .

FROM $BUILD_FROM

RUN apk add curl
RUN apk add chromium

WORKDIR /
COPY --from=builder /usr/bin/solar-assistant-browser-automation /
COPY run.sh .

EXPOSE 8080

HEALTHCHECK --interval=5s --timeout=5s --start-period=10s --retries=3 CMD curl --fail http://localhost:8080/health || exit 1

CMD [ "./run.sh" ]
