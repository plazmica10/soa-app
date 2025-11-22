# Makefile for generating protobuf files

.PHONY: proto clean

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		protos/stakeholders.proto protos/follower.proto

clean:
	rm -f protos/*.pb.go
