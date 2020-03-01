FROM alpine:3.11 as builder

RUN apk add make gcc musl-dev linux-headers libpcap-dev clang

ARG MASSCAN_VERSION

RUN wget https://github.com/robertdavidgraham/masscan/archive/${MASSCAN_VERSION}.tar.gz
RUN tar xf ${MASSCAN_VERSION}.tar.gz
RUN mv masscan-${MASSCAN_VERSION} masscan
WORKDIR /masscan
RUN make -j


FROM alpine:3.11

ARG MASSCAN_VERSION
ARG BUILD_DATE

RUN apk add --no-cache libpcap-dev ca-certificates
COPY --from=builder /masscan/bin/masscan /usr/bin/masscan

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="masscan" \
      org.label-schema.description="TCP port scanner, spews SYN packets asynchronously, scanning entire Internet in under 5 minutes." \
      org.label-schema.url="https://github.com/robertdavidgraham/masscan" \
      org.label-schema.vcs-ref=${MASSCAN_VERSION} \
      org.label-schema.vcs-url="https://github.com/robertdavidgraham/masscan" \
      org.label-schema.schema-version="1.0"

ENTRYPOINT ["/usr/bin/masscan"]
