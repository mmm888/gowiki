## all
all: run

## Setup
setup:
	go get github.com/Songmu/make2help/cmd/make2help

## Install Dependencies
deps: setup
	glide install

## Update Dependencies
update: setup
	glide update

## Test
test: deps
	go test $$(glide novendor)

## Run
run: deps
	go run main.go type.go utils.go git.go

## Help
help:
	@make2help $(MAKEFILE_LIST)

.PHONY: all setup deps update test run help
