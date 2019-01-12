ROOT=$(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
IMAGE=fopina/scanme:masscan

all: build push

build:
	@docker build -t $(IMAGE) .

push: 
	@docker push $(IMAGE)

.PHONY: all build push