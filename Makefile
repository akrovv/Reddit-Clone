build:
	mkdir api && go build -o ./api ./...

test-api:
	go test --cover ./... | grep -v 'no test files'

test-permissions:
	./tests_permissions.sh

lint:
	golangci-lint run ./...