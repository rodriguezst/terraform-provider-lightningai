default: build

build:
	go build -o terraform-provider-lightningai .

install: build
	mkdir -p ~/.terraform.d/plugins/lightningai/lightning/1.0.0/$$(go env GOOS)_$$(go env GOARCH)/
	cp terraform-provider-lightningai ~/.terraform.d/plugins/lightningai/lightning/1.0.0/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -f terraform-provider-lightningai

.PHONY: build install test fmt vet lint clean
