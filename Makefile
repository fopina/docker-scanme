bin/fakemasscan: fakemasscan.go
	go build -o bin/fakemasscan -ldflags '-w -s' fakemasscan.go

bin/scanme: scanme.go
	go build -o bin/scanme -ldflags '-w -s' scanme.go

.PHONE: test
test: bin/fakemasscan bin/scanme
	bin/scanme -path bin/fakemasscan -sleep 0 -show 45.33.32.156

.PHONE: longtest
longtest: bin/fakemasscan bin/scanme
	$(eval export FAKEMASSCAN := $(shell mktemp))
	bin/fakemasscan -setup 1,2,3 2,3 3,4
	bin/scanme -path bin/fakemasscan -sleep 1 45.33.32.156

.PHONY: build
build: test
	@docker build -t fopina/scanme .

.PHONY: push
push: 
	@docker push fopina/scanme:latest

.PHONY: hub
hub: build push

.PHONY: clean
clean:
	rm -f bin/*
