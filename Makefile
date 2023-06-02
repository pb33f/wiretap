build: build-ui build-daemon

build-ui:
	@cd ui && yarn install && yarn build

build-daemon:
	@go build -o bin/wiretap