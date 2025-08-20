GO_BIN := $(shell go env GOPATH)/bin
SWAG := $(GO_BIN)/swag
SWAG_INSTALL := go install github.com/swaggo/swag

.PHONY: ensure-tools swag swagger-clean run doctor

ensure-tools:
	@if [ ! -x "$(SWAG)" ]; then \
		echo "Installing swag..."; \
		$(SWAG_INSTALL); \
	else \
		echo "swag found at $(SWAG)"; \
	fi

# swagger docs
swag: ensure-tools swag-clean
	$(SWAG) init \
		-g main.go \
		-d ./cmd/server,./internal \
		-o ./docs

# swagger clean
swag-clean:
	rm -rf ./docs

# swagger fmt
swag-fmt:
	$(SWAG) fmt

# first swagger & run
run: swag swag-fmt
	go run ./cmd/server

# self-check
doctor:
	@echo "GO_BIN: $(GO_BIN)"
	@echo -n "swag version: "; if [ -x "$(SWAG)" ]; then $(SWAG) --version; else echo "not installed"; fi
	@echo "main.go exists?"; [ -f ./cmd/server/main.go ] && echo "Yes" || (echo "No"; exit 1)
