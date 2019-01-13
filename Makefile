ROOT=$(abspath $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST))))))
IMAGE=fopina/scanme

all: build push

$(BINDIR)bin/fakemasscan: fakemasscan.go
	go build -o $(BINDIR)bin/fakemasscan -ldflags '-w -s' fakemasscan.go

$(BINDIR)bin/scanme: scanme.go
	go build -o $(BINDIR)bin/scanme -ldflags '-w -s' scanme.go

test: $(BINDIR)bin/fakemasscan $(BINDIR)bin/scanme
	$(BINDIR)bin/scanme -path $(BINDIR)bin/fakemasscan -sleep 0 -show 45.33.32.156

longtest: $(BINDIR)bin/fakemasscan $(BINDIR)bin/scanme
	$(eval export FAKEMASSCAN := $(shell mktemp))
	$(BINDIR)bin/fakemasscan -setup 1,2,3 2,3 3,4
	$(BINDIR)bin/scanme -path $(BINDIR)bin/fakemasscan -sleep 1 45.33.32.156

dockertest:
	docker run -v $(ROOT):/app:ro -w /app golang:1.11-alpine3.8 sh -c 'apk add make && make test BINDIR=/appbin/'

build: dockertest
	docker build -t $(IMAGE) .

push: 
	docker push $(IMAGE):latest

travis-tag: build
	docker tag $(IMAGE) $(IMAGE):latest
	docker tag $(IMAGE) $(IMAGE):$(TAG)
	docker push $(IMAGE):latest
	docker push $(IMAGE):$(TAG)

travis-dev: build
	docker tag $(IMAGE) $(IMAGE):dev
	docker push $(IMAGE):dev

clean:
	rm -f bin/*

.PHONY: all clean push build longtest test dockertest