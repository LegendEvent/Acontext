GO_DIR := ./go
GO_BIN := $(shell go env GOPATH)/bin
SWAG := $(GO_BIN)/swag
SWAG_INSTALL := go install github.com/swaggo/swag

.PHONY: ensure-tools swag swag-clean swag-fmt run doctor

# check tools
ensure-tools:
	@if [ ! -x "$(SWAG)" ]; then \
		echo "Installing swag..."; \
		(cd $(GO_DIR) && $(SWAG_INSTALL)); \
	else \
		echo "swag found at $(SWAG)"; \
	fi

# swagger docs
swag: ensure-tools swag-clean
	cd $(GO_DIR) && $(SWAG) init \
		-g main.go \
		-d ./cmd/server,./internal \
		-o ./docs

# swagger clean
swag-clean:
	rm -rf $(GO_DIR)/docs

# swagger fmt
swag-fmt:
	cd $(GO_DIR) && $(SWAG) fmt

# first swagger & run
run: swag swag-fmt
	@set -a; . ../.env; set +a; \
	cd $(GO_DIR) && go run ./cmd/server

# self-check
doctor:
	@echo "GO_BIN: $(GO_BIN)"
	@echo -n "swag version: "; if [ -x "$(SWAG)" ]; then $(SWAG) --version; else echo "not installed"; fi
	@echo "main.go exists?"; [ -f $(GO_DIR)/cmd/server/main.go ] && echo "Yes" || (echo "No"; exit 1)
