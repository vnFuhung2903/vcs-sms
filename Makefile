test:
	go test -coverprofile=coverage.out $(shell go list ./... | grep -vE '/logs|/mocks|/cmd|/data')
	go tool cover -func=coverage.out