FROM golang:1.16-alpine3.13

RUN apk add --no-cache \
  chromium \
  chromium-chromedriver

ENV CGO_ENABLED=0

COPY . /go/src/github.com/rm-jooho/webrtc
WORKDIR /go/src/github.com/rm-jooho/webrtc/e2e

CMD ["go", "test", "-tags=e2e", "-v", "."]
