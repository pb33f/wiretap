
build: build-ui build-daemon

build-ui:
	@cd ui && yarn install
	@cd ui && yarn build

build-daemon:
	@go build -o bin/wiretap

run:
	@go run wiretap.go