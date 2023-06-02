build:
	@cd ui && yarn install
	@cd ui && yarn build
	@go build -o bin/wiretap

run:
	@go run wiretap.go