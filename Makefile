build: build-ui build-daemon

build-ui:
	@cd ui && npm install && npm run build

build-daemon:
	@go build -o bin/wiretap