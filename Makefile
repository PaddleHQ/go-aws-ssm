
ifndef UNIQUE_BUILD_ID
	UNIQUE_BUILD_ID=latest
endif

.PHONY: build
.SILENT: help
help: ## Show this help message
	set -x

	echo ""
	echo "Available targets:"
	grep ':.* ##\ ' ${MAKEFILE_LIST} | awk '{gsub(":[^#]*##","\t"); print}' | column -t -c 2 -s $$'\t' | sort

# Build the container
build: ## Build the container
	docker build --tag=go-aws-ssm:$(UNIQUE_BUILD_ID) .

unit-test:
	docker run go-aws-ssm:$(UNIQUE_BUILD_ID) /bin/bash -c "go test ./... -v"

vet:
	docker run go-aws-ssm:$(UNIQUE_BUILD_ID) /bin/bash -c  "go vet ./..."

lint:
	docker run go-aws-ssm:$(UNIQUE_BUILD_ID) /bin/bash -c  "golint -set_exit_status ./..."
