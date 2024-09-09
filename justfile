# show help by default
default:
    @just --list --justfile {{ justfile() }}

# update go deps
update *flags:
    go get {{ flags }} ./cmd/honoroit
    go mod tidy
    go mod vendor

# run linter
lint: && swagger
    golangci-lint run ./...

# automatically fix liter issues
lintfix: && swaggerfix
    golangci-lint run --fix ./...

# generate mocks
mocks:
    @mockery --all --inpackage --testonly --exclude vendor

# generate swagger docks
swagger:
    @swag init --parseDependency --dir ./cmd/honoroit,./

# automatically fix swagger issues
swaggerfix: && swagger
    @swag fmt --dir ./cmd/honoroit,./

# run unit tests
test packages="./...":
    @go test -cover -coverprofile=cover.out -coverpkg={{ packages }} -covermode=set {{ packages }}
    @go tool cover -func=cover.out
    -@rm -f cover.out

# run app
run:
    @go run ./cmd/honoroit

install:
    @CGO_ENABLED=0 go install -ldflags '-extldflags "-static"' -tags timetzdata,goolm -v ./cmd/honoroit

# build app
build:
    @CGO_ENABLED=0 go build -ldflags '-extldflags "-static"' -tags timetzdata,goolm -v ./cmd/honoroit
