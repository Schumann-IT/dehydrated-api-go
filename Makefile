.PHONY: gen
gen:
	go generate ./...

.PHONY: test
test:
	go test ./...