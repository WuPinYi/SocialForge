.PHONY: generate build run clean

generate:
	go generate ./ent
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/ocs.proto

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -rf bin/
	rm -f proto/*.pb.go 