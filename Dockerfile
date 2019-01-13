FROM golang:1.11-alpine3.8 as builder1

WORKDIR /go/src/app
COPY scanme.go /go/src/app
RUN go build scanme.go

FROM fopina/scanme:masscan as builder2

# nothing to do, just speed up travis-ci build
# image built in orphan branch "masscan"

FROM alpine:3.8

RUN apk add --no-cache libpcap-dev ca-certificates
COPY --from=builder1 /go/src/app/scanme /usr/bin/scanme
COPY --from=builder2 /usr/bin/masscan /usr/bin/masscan

ENTRYPOINT ["scanme"]