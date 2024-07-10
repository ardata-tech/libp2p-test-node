GOPATH:=$(shell go env GOPATH)

.PHONY: init
init:
	@go mod tidy && go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: update
update:
	@go get -u

.PHONY: tidy
tidy:
	@go mod tidy && go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: build
build:
	@go mod tidy && go install github.com/swaggo/swag/cmd/swag@latest && go build -o libp2p-node *.go

.PHONY: generate-swagger
generate-swagger:
	@swag init -g api/*.go

.PHONY: test
test:
	@go test -v ./... -cover

.PHONY: docker
docker:
	@docker build -t libp2p-node:latest .

.PHONY: run
run:
	@go run *.go


.PHONY: build-run
build-run:
	@go mod tidy && go install github.com/swaggo/swag/cmd/swag@latest && go build -o libp2p-node *.go && ./libp2p-node