build: build-ui build-daemon

build-ui:
	@cd ui && npm ci && npm run build

build-daemon:
	@go build -o bin/wiretap