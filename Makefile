ROOT=$(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
IMAGE=fopina/scanme:masscan

build:
	@docker build -t $(IMAGE) .

push: 
	@docker push $(IMAGE)

all: build push

.PHONY: build push hub