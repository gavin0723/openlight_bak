# Protobuf makefile

.SILENT:
.PHONY: build clean

ProtoFiles := ./spec/build.proto ./spec/runner.proto

build: clean
	mkdir ../protoc-gen-go
	protoc -I . --go_out=../protoc-gen-go $(ProtoFiles)
clean:
	rm -rf ../protoc-gen-go
