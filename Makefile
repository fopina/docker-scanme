ROOT=$(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
IMAGE=fopina/scanme:masscan

all: build push

build:
	@docker build -t $(IMAGE) \
				  --build-arg MASSCAN_VERSION=8189d513fb9ecd8333a8be7475044d03fb029318 \
				  --build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
				  .

push: 
	@docker push $(IMAGE)

.PHONY: all build push
