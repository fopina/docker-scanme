FROM golang:1.11-alpine3.8 as builder1

WORKDIR /go/src/app
COPY scanme.go /go/src/app
RUN go build scanme.go

FROM alpine:3.8 as builder2

RUN apk add make gcc musl-dev linux-headers libpcap-dev clang

ARG MASSCAN_VERSION=master

RUN wget https://github.com/robertdavidgraham/masscan/archive/${MASSCAN_VERSION}.tar.gz
RUN tar xvf ${MASSCAN_VERSION}.tar.gz
RUN mv masscan-${MASSCAN_VERSION} masscan
WORKDIR /masscan
RUN make -j


FROM alpine:3.8

RUN apk add --no-cache libpcap-dev
COPY --from=builder1 /go/src/app/scanme /usr/bin/scanme
COPY --from=builder2 /masscan/bin/masscan /usr/bin/masscan

ENTRYPOINT ["scanme"]