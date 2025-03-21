.PHONY: proto
proto:
	go generate ./...

.PHONY: test
test:
	go test ./... 