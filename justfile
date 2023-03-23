CI_REGISTRY_IMAGE := env_var_or_default("CI_REGISTRY_IMAGE", "registry.gitlab.com/etke.cc/honoroit")
REGISTRY_IMAGE := env_var_or_default("REGISTRY_IMAGE", "registry.etke.cc/etke.cc/honoroit")
CI_COMMIT_TAG := if env_var_or_default("CI_COMMIT_TAG", "main") == "main" { "latest" } else { env_var_or_default("CI_COMMIT_TAG", "latest") }

# show help by default
default:
    @just --list --justfile {{ justfile() }}

# update go deps
update:
    go get ./cmd
    go mod tidy
    go mod vendor

# run linter
lint:
    golangci-lint run ./...

# automatically fix liter issues
lintfix:
    golangci-lint run --fix ./...

# generate mocks
mocks:
    @rm -rf mocks
    @mockery --all

# run unit tests
test:
    @go test -coverprofile=cover.out ./...
    @go tool cover -func=cover.out
    -@rm -f cover.out

# run app
run:
    @go run ./cmd

# build app
build:
    go build -v -o honoroit ./cmd

# docker login
login:
    @docker login -u gitlab-ci-token -p $CI_JOB_TOKEN $CI_REGISTRY

# docker build
docker:
    docker buildx create --use
    docker buildx build --platform linux/arm64/v8,linux/amd64 --push -t {{ CI_REGISTRY_IMAGE }}:{{ CI_COMMIT_TAG }} -t {{ REGISTRY_IMAGE }}:{{ CI_COMMIT_TAG }} .
