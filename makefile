# The openlight cli makefile

.PHONY: build go

Targets := ./cli/op

build: go

go:
	go install $(Targets)

