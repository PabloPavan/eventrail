.PHONY: vet lint check

vet:
	go vet ./...

lint:
	golangci-lint run

check: vet lint
