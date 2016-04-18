APPENV ?= testenv
PROJECT := $(shell basename $$PWD)
REV ?= latest

all: build

clean:
	rm -fr target bin pkg

fmt:
	@gofmt -w ./

deps:
	docker-compose up -d
	docker run --link $(PROJECT)_postgres_1:postgres aanand/wait

build: deps $(APPENV)
	docker run \
		--link $(PROJECT)_postgres_1:postgres \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-e PROJECT=github.com/opsee/$(PROJECT) \
		-v `pwd`:/gopath/src/github.com/opsee/$(PROJECT) \
		quay.io/opsee/build-go:16
		docker build -t quay.io/opsee/$(PROJECT):$(REV) .

run: deps $(APPENV)
	docker run \
		--link $(PROJECT)_postgres_1:postgres \
		--env-file ./$(APPENV) \
		-e AWS_DEFAULT_REGION \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-p 9104:9104 \
		--rm \
		quay.io/opsee/$(PROJECT):$(REV)

.PHONY: docker run migrate clean all