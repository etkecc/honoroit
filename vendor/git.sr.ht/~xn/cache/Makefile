update:
	go get .
	go mod tidy

lint:
	golangci-lint run ./...

lintfix:
	golangci-lint run --fix ./...

test:
	@go test ${BUILDFLAGS} -coverprofile=cover.out ./...
	@go tool cover -func=cover.out
	-@rm -f cover.out

bench:
	@go test -bench=. ./...
