# The openlight cli makefile

.PHONY: build go godeps

Targets := ./cli/op

build: go

go:
	go install $(Targets)

godeps:
	godep save $(Targets)

