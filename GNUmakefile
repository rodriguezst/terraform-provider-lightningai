default: build

build:
	go build -o terraform-provider-lightningai .

install: build
	mkdir -p ~/.terraform.d/plugins/rodriguezst/lightningai/1.0.1/$$(go env GOOS)_$$(go env GOARCH)/
	cp terraform-provider-lightningai ~/.terraform.d/plugins/rodriguezst/lightningai/1.0.1/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./...

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

generate:
	go generate ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -f terraform-provider-lightningai terraform-provider-lightningai.exe

.PHONY: build install test testacc generate fmt vet lint clean
