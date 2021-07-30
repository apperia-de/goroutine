final-check: build mod-tidy test test-code-coverage lint

build:
	go build

test:
	go test

test-verbose:
	go test -v

lint:
	golangci-lint run

test-code-coverage:
	go test -cover

show-code-coverage:
	go test -coverprofile=c.out && go tool cover -html=c.out && rm c.out

create-code-coverage-report:
	go test -coverprofile=c.out && go tool cover -html=c.out -o coverage.html
	rm c.out

mod-tidy:
	go mod tidy