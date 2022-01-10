BUILDFLAGS = -buildmode=pie
### CI vars
CI_LOGIN_COMMAND = @echo "Not a CI, skip login"
CI_REGISTRY_IMAGE ?= registry.gitlab.com/etke.cc/honoroit
CI_COMMIT_TAG ?= latest
# for main branch it must be set explicitly
ifeq ($(CI_COMMIT_TAG), main)
CI_COMMIT_TAG = latest
endif
# login command
ifdef CI_JOB_TOKEN
CI_LOGIN_COMMAND = @docker login -u gitlab-ci-token -p $(CI_JOB_TOKEN) $(CI_REGISTRY)
endif

# update go dependencies
update:
	go get ./cmd
	go mod tidy
	go mod vendor

mock:
	-@rm -rf mocks
	@mockery --all

# run linter
lint:
	golangci-lint run ./...

# run linter and fix issues if possible
lintfix:
	golangci-lint run --fix ./...

# run unit tests
test:
	@go test ${BUILDFLAGS} -coverprofile=cover.out ./...
	@go tool cover -func=cover.out
	-@rm -f cover.out

# run honoroit, note: make doesn't understand exit code 130 and sets it == 1
run:
	@go run ${BUILDFLAGS} ./cmd || exit 0

# build honoroit
build:
	go build ${BUILDFLAGS} -v -o honoroit -ldflags "-X main.version=${CI_COMMIT_TAG}" ./cmd

# CI: docker login
login:
	@echo "trying to login to docker registry..."
	$(CI_LOGIN_COMMAND)

# docker build
docker:
	docker buildx create --use
	docker buildx build --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --push -t ${CI_REGISTRY_IMAGE}:${CI_COMMIT_TAG} .
