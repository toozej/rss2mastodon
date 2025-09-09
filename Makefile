# Set sane defaults for Make
SHELL = bash
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

# Set default goal such that `make` runs `make help`
.DEFAULT_GOAL := help

# Build info
BUILDER = $(shell whoami)@$(shell hostname)
NOW = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Version control
VERSION = $(shell git describe --tags --dirty --always)
COMMIT = $(shell git rev-parse --short HEAD)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

# Linker flags
PKG = $(shell head -n 1 go.mod | cut -c 8-)
VER = $(PKG)/pkg/version
LDFLAGS = -s -w \
	-X $(VER).Version=$(or $(VERSION),unknown) \
	-X $(VER).Commit=$(or $(COMMIT),unknown) \
	-X $(VER).Branch=$(or $(BRANCH),unknown) \
	-X $(VER).BuiltAt=$(NOW) \
	-X $(VER).Builder=$(BUILDER)
	
OS = $(shell uname -s)
ifeq ($(OS), Linux)
	OPENER=xdg-open
else
	OPENER=open
endif

.PHONY: all vet test build verify run up down local local-vet local-test local-cover local-run local-release-test local-release local-sign local-verify local-release-verify install get-cosign-pub-key docker-login pre-commit-install pre-commit-run pre-commit pre-reqs update-golang-version upload-secrets-to-gh upload-secrets-envfile-to-1pass docs clean help

all: vet pre-commit clean test build verify run ## Run default workflow via Docker
local: local-update-deps local-vendor local-vet pre-commit clean local-test local-cover local-build local-sign local-verify local-run ## Run default workflow using locally installed Golang toolchain
local-release-verify: local-release local-sign local-verify ## Release and verify using locally installed Golang toolchain
pre-reqs: pre-commit-install ## Install pre-commit hooks and necessary binaries

vet: ## Run `go vet` in Docker
	docker build --target vet -f $(CURDIR)/Dockerfile.distroless -t toozej/rss2mastodon:latest . 

test: ## Run `go test` in Docker
	docker build --progress=plain --target test -f $(CURDIR)/Dockerfile.distroless -t toozej/rss2mastodon:latest . 

build: ## Build Docker image, including running tests
	docker build -f $(CURDIR)/Dockerfile.distroless -t toozej/rss2mastodon:latest .

get-cosign-pub-key: ## Get rss2mastodon Cosign public key from GitHub
	test -f $(CURDIR)/rss2mastodon.pub || curl --silent https://raw.githubusercontent.com/toozej/rss2mastodon/main/rss2mastodon.pub -O

verify: get-cosign-pub-key ## Verify Docker image with Cosign
	cosign verify --key $(CURDIR)/rss2mastodon.pub toozej/rss2mastodon:latest

run: ## Run built Docker image
	docker run --rm --name rss2mastodon --env-file $(CURDIR)/.env toozej/rss2mastodon:latest

up: test build ## Run Docker Compose project with build Docker image
	docker compose -f docker-compose.yml down --remove-orphans
	docker compose -f docker-compose.yml pull
	docker compose -f docker-compose.yml up -d

down: ## Stop running Docker Compose project
	docker compose -f docker-compose.yml down --remove-orphans

local-update-deps: ## Run `go get -t -u ./...` to update Go module dependencies
	go get -t -u ./...

local-vet: ## Run `go vet` using locally installed golang toolchain
	go vet $(CURDIR)/...

local-vendor: ## Run `go mod tidy & vendor` using locally installed golang toolchain
	go mod tidy
	go mod vendor

local-test: ## Run `go test` using locally installed golang toolchain
	go test -coverprofile c.out -v $(CURDIR)/...
	@echo -e "\nStatements missing coverage"
	@grep -v -e " 1$$" c.out

local-cover: ## View coverage report in web browser
	go tool cover -html=c.out

local-build: ## Run `go build` using locally installed golang toolchain
	CGO_ENABLED=0 go build -o $(CURDIR)/out/ -ldflags="$(LDFLAGS)"

local-run: ## Run locally built binary
	if test -e $(CURDIR)/.env; then \
		export `cat $(CURDIR)/.env | xargs` && $(CURDIR)/out/rss2mastodon; \
	else \
		echo "No environment variables found at $(CURDIR)/.env. Cannot run."; \
	fi

local-release-test: ## Build assets and test goreleaser config using locally installed golang toolchain and goreleaser
	goreleaser check
	goreleaser build --rm-dist --snapshot

local-release: local-test docker-login ## Release assets using locally installed golang toolchain and goreleaser
	if test -e $(CURDIR)/rss2mastodon.key && test -e $(CURDIR)/.env; then \
		export `cat $(CURDIR)/.env | xargs` && goreleaser release --rm-dist; \
	else \
		echo "no cosign private key found at $(CURDIR)/rss2mastodon.key. Cannot release."; \
	fi

local-sign: local-test ## Sign locally installed golang toolchain and cosign
	if test -e $(CURDIR)/rss2mastodon.key && test -e $(CURDIR)/.env; then \
		export `cat $(CURDIR)/.env | xargs` && cosign sign-blob --key=$(CURDIR)/rss2mastodon.key --output-signature=$(CURDIR)/rss2mastodon.sig $(CURDIR)/out/rss2mastodon; \
	else \
		echo "no cosign private key found at $(CURDIR)/rss2mastodon.key. Cannot release."; \
	fi

local-verify: get-cosign-pub-key ## Verify locally compiled binary
	# cosign here assumes you're using Linux AMD64 binary
	cosign verify-blob --key $(CURDIR)/rss2mastodon.pub --signature $(CURDIR)/rss2mastodon.sig $(CURDIR)/out/rss2mastodon

install: local-build local-verify ## Install compiled binary to local machine
	sudo cp $(CURDIR)/out/rss2mastodon /usr/local/bin/rss2mastodon
	sudo chmod 0755 /usr/local/bin/rss2mastodon

upload-secrets-to-gh: ## Upload secrets from .env file to GitHub Actions Secrets + Dependabot
	$(CURDIR)/scripts/upload_secrets_to_github.sh rss2mastodon 

upload-secrets-envfile-to-1pass: ## Upload secrets and .env file to 1Password
	$(CURDIR)/scripts/upload_secrets_to_1password secrets rss2mastodon
	$(CURDIR)/scripts/upload_secrets_to_1password envfile rss2mastodon

docker-login: ## Login to Docker registries used to publish images to
	if test -e $(CURDIR)/.env; then \
		export `cat $(CURDIR)/.env | xargs`; \
		echo $${DOCKERHUB_TOKEN} | docker login docker.io --username $${DOCKERHUB_USERNAME} --password-stdin; \
		echo $${QUAY_TOKEN} | docker login quay.io --username $${QUAY_USERNAME} --password-stdin; \
		echo $${GITHUB_GHCR_TOKEN} | docker login ghcr.io --username $${GITHUB_USERNAME} --password-stdin; \
	else \
		echo "No container registry credentials found, need to add them to ./.env. See README.md for more info"; \
	fi

pre-commit: pre-commit-install pre-commit-run ## Install and run pre-commit hooks

pre-commit-install: ## Install pre-commit hooks and necessary binaries
	# golangci-lint
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	# goimports
	go install golang.org/x/tools/cmd/goimports@latest
	# gosec
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	# staticcheck
	go install honnef.co/go/tools/cmd/staticcheck@latest
	# go-critic
	go install github.com/go-critic/go-critic/cmd/gocritic@latest
	# structslop
	# go install github.com/orijtech/structslop/cmd/structslop@latest
	# shellcheck
	command -v shellcheck || sudo dnf install -y ShellCheck || sudo apt install -y shellcheck
	# checkmake
	go install github.com/checkmake/checkmake/cmd/checkmake@latest
	# goreleaser
	go install github.com/goreleaser/goreleaser/v2@latest
	# syft
	command -v syft || curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
	# cosign
	go install github.com/sigstore/cosign/cmd/cosign@latest
	# go-licenses
	go install github.com/google/go-licenses@latest
	# go vuln check
	go install golang.org/x/vuln/cmd/govulncheck@latest
	# air
	go install github.com/air-verse/air@latest
	# graphviz for dot
	command -v dot || brew install graphviz || sudo apt install -y graphviz || sudo dnf install -y graphviz
	# install and update pre-commits
	# determine if on Debian 12 and if so use pip to install more modern pre-commit version
	grep --silent "VERSION=\"12 (bookworm)\"" /etc/os-release && apt install -y --no-install-recommends python3-pip && python3 -m pip install --break-system-packages --upgrade pre-commit || echo "OS is not Debian 12 bookworm"
	command -v pre-commit || brew install pre-commit || sudo dnf install -y pre-commit || sudo apt install -y pre-commit
	pre-commit install
	pre-commit autoupdate

pre-commit-run: ## Run pre-commit hooks against all files
	pre-commit run --all-files
	# manually run the following checks since their pre-commits aren't working or don't exist
	go-licenses report github.com/toozej/rss2mastodon/cmd/rss2mastodon
	govulncheck ./...

update-golang-version: ## Update to latest Golang version across the repo
	@VERSION=`curl -s "https://go.dev/dl/?mode=json" | jq -r '.[0].version' | sed 's/go//' | cut -d '.' -f 1,2`; \
	$(CURDIR)/scripts/update_golang_version.sh $$VERSION

docs: ## Serve Go documentation
	@echo "Starting Go documentation server on localhost"
	@echo "Use Ctrl+C to stop the server"
	go doc -http

clean: ## Remove any locally compiled binaries
	rm -f $(CURDIR)/out/rss2mastodon

help: ## Display help text
	@grep -E '^[a-zA-Z_-]+ ?:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
