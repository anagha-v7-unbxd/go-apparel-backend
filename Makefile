SHELL := /bin/bash

# Config
# 
# Change these properties based on the project
# -------------------------------------------------------------------------------------------------------------
proj_name := gostarter
bin_name := $(proj_name).bin
main_pkg := cmd/app
latest := latest
ecr_url := 012629307706.dkr.ecr.us-east-1.amazonaws.com
# -------------------------------------------------------------------------------------------------------------

ifdef SKAFFOLD_DEFAULT_REPO
ecr_url := $(SKAFFOLD_DEFAULT_REPO)
endif
ifdef ECR_URL
ecr_url := $(ECR_URL)
endif

# Check dependencies
# -- Go Binary
bin_go := $(shell command -v go 2> /dev/null)
ifndef bin_go
$(error Missing `go` binary from $PATH)
endif


# -- Docker
docker_bin := $(shell command -v docker 2> /dev/null)
ifndef docker_bin
$(error Missing `docker` binary from $PATH)
endif

# gcc-11 only required for CGO builds; skip on macOS / when cross-compiling
gcc_11_bin := $(shell command -v gcc-11 2> /dev/null)
ifdef CGO_ENABLED
ifndef gcc_11_bin
$(warning gcc-11 not found - set CGO_ENABLED=0 or install gcc-11 for CGO builds)
endif
endif


# Extract project info
# project
mod_name := $(shell head -1 go.mod | cut -d' ' -f2)
proj_dir := $(shell pwd)
# git
git_hash := $(shell git rev-parse --short=16 HEAD)
git_branch := $(shell git rev-parse --abbrev-ref HEAD)
git_stat := $(shell if [ "$$(git diff --stat)" != '' ]; then echo "dirty"; else echo "clean"; fi)
git_tag := $(shell tag=$$(git describe --tags --abbrev=0 2>/dev/null); if [ $$? -ne 0 ]; then echo "latest"; else echo "$$tag"; fi)
# build date
build_date := $(shell date +%s)
build_version := ${BUILD_VERSION}
skaffold_bin := $(shell command -v skaffold 2> /dev/null)
helm_bin := $(shell command -v helm 2> /dev/null)

ifeq ($(build_version),)
build_version := $(git_tag)
endif

ldflags := -X $(mod_name)/cmd/ldflags.Module=$(mod_name) \
			-X $(mod_name)/cmd/ldflags.GitHash=$(git_hash) \
			-X $(mod_name)/cmd/ldflags.GitBranch=$(git_branch) \
			-X $(mod_name)/cmd/ldflags.GitStat=$(git_stat) \
			-X $(mod_name)/cmd/ldflags.GitTag=$(git_tag) \
			-X $(mod_name)/cmd/ldflags.BuildDate=$(build_date) \
			-X $(mod_name)/cmd/ldflags.Version=$(build_version)

ldflags_cmd := -ldflags="$(ldflags)"

default: help

## info: prints the defailts about the project
info:
	@echo "--------------------------------------------"
	@echo "               $(shell echo $(proj_name) | tr  '[:lower:]' '[:upper:]') MAKE FILE              "
	@echo "--------------------------------------------"
	@printf "Bin:			%s\n" "$(bin_name)"
	@printf "Module:			%s\n" "$(mod_name)"
	@printf "Proj Dir:		%s\n" "$(proj_dir)"
	@printf "Proj Name:		%s\n" "$(proj_name)"
	@printf "Main Pkg:		%s\n" "$(main_pkg)"
	@printf "Git Hash:		%s\n" "$(git_hash)"
	@printf "Git Branch:		%s\n" "$(git_branch)"
	@printf "Git Stat:		%s\n" "$(git_stat)"
	@printf "Git Tag:		%s\n" "$(git_tag)"
	@printf "Build Date:		%s\n" "$(build_date)"
	@printf "Build Version:		%s\n" "$(build_version)"
	@echo "--------------------------------------------"

info-ldflags:
	@echo " > LD FLags"
	@echo "--------------------------------------------"
	@echo "$(ldflags)"
	@echo "--------------------------------------------"
.PHONY: print info info-ldflags

goclean:
	@echo " > Delete: rm -f $(proj_dir)/bin/$(bin_name) 2>/dev/null"
	@rm -f $(proj_dir)/bin/$(bin_name) 2>/dev/null
goensure:
	@echo " > Ensure: go mod tidy"
	@go mod tidy
gotest:
	@echo " > Run Test: go test ./..."
	@echo " > --------- "
	@go test ./...
	@echo " > --------- "
movebin:
	@echo " > Make Dir:  mkdir -p $(proj_dir)/bin"
	@mkdir -p ./bin
	@echo " > Move: mv $(bin_name) $(proj_dir)/bin"
	@mv $(bin_name) $(proj_dir)/bin
buildpost:
	@echo " > --------------------------------------------"
	@echo " > Go Binary Generated "
	@echo " > --------------------------------------------"
	
## gobuild: builds the binary for linux
gobuild:
	@-$(MAKE) --no-print-directory info info-ldflags goclean goensure gotest
	@echo " > Build Binary: env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(bin_name) $(ldflags_cmd) $(mod_name)/$(main_pkg)"
	@env GOSUMDB=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(bin_name) $(ldflags_cmd) $(mod_name)/$(main_pkg)
	@-$(MAKE) --no-print-directory movebin buildpost

## gobuildmac: builds the binary for mac-arm
gobuildmac:
	@-$(MAKE) --no-print-directory goclean goensure gotest
	@echo " > Build Binary: env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(bin_name) $(ldflags_cmd) $(mod_name)/$(main_pkg)"
	@env GOSUMDB=off CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(bin_name) $(ldflags_cmd) $(mod_name)/$(main_pkg)
	@-$(MAKE) --no-print-directory movebin buildpost
	
.PHONY: goclean goensure gotest movebin buildpost gobuild gobuildmac

docker-build:
	@echo " > Build: docker build -t $(proj_name):latest $(proj_dir)"
	@docker build -t $(proj_name):latest $(proj_dir)

docker-push:
	@echo " > Docker Push: pushing image : [ $(ecr_url)/$(proj_name):latest ]"
	@docker tag $(proj_name):latest $(ecr_url)/$(proj_name):latest
	@docker push $(ecr_url)/$(proj_name):latest
ifneq ($(build_version), $(latest))
	@echo " > Docker Push: pushing image : [ $(ecr_url)/$(proj_name):$(build_version) ]"
	@docker tag $(proj_name):$(git_tag) $(ecr_url)/$(proj_name):$(build_version)
	@docker push $(ecr_url)/$(proj_name):$(build_version)
endif

docker-post:
	@echo " > --------------------------------------------"
	@echo " > Docker Image Generated "
	@echo " > --------------------------------------------"

## docker: builds docker image for the project using Dockerfile
docker:
	@-$(MAKE) --no-print-directory docker-build docker-push docker-post

## docker-run: runs the docker image in your local Docker environment
docker-run:
	@-$(MAKE) --no-print-directory build docker
	@echo " > Running: docker run $(proj_name):latest"
	@docker run --env SERVICE=$(proj_name) $(proj_name):latest

.PHONY: docker-build docker-push docker-post docker docker-run 

## build: builds binary for linux and generates docker image
build:
	@-$(MAKE) --no-print-directory gobuild docker

## build-mac: builds binary for mac-arm and generates docker image
build-mac:
	@-$(MAKE) --no-print-directory gobuildmac docker

.PHONY: build build-mac

skaffold-check: skaffold-exists helm-exists
skaffold-exists: ; @which skaffold > /dev/null
helm-exists: ; @which helm > /dev/null

## skaffold: run local skaffold in dev environment
skaffold: skaffold-check
	@-$(MAKE) --no-print-directory gobuild
	@skaffold dev

## skaffold-run: run skaffold in normal mode
skaffold-run: skaffold-check
	@-$(MAKE) --no-print-directory build
	@skaffold run

.PHONY: skaffold-check skaffold-exists helm-exists skaffold skaffold-run

## bin-run: runs the binary created using the runner-script
bin-run: export SERVICE = $(proj_name)
bin-run:
	@-$(MAKE) --no-print-directory gobuild 
	@./scripts/run.sh
.PHONY: bin-run

## help: print this help message
help: Makefile
	@echo " ---------------------------------------"
	@echo "               $(shell echo $(proj_name) | tr  '[:lower:]' '[:upper:]') MAKE FILE              "
	@echo " ---------------------------------------"
	@echo " Choose a command run:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

.PHONY: help
