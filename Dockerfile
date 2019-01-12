FROM alpine:3.8 as builder

RUN apk add make gcc musl-dev linux-headers libpcap-dev clang

ARG MASSCAN_VERSION=master

RUN wget https://github.com/robertdavidgraham/masscan/archive/${MASSCAN_VERSION}.tar.gz
RUN tar xvf ${MASSCAN_VERSION}.tar.gz
RUN mv masscan-${MASSCAN_VERSION} masscan
WORKDIR /masscan
RUN make -j


FROM alpine:3.8

RUN apk add --no-cache libpcap-dev ca-certificates
COPY --from=builder /masscan/bin/masscan /usr/bin/masscan

ENTRYPOINT ["/usr/bin/masscan"]