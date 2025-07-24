# Define binary names
CONTROLLER_BIN=controller
RUNNER_BIN=runner
BUILD_BIN=plumbops-build

# Define build flags
BUILD_FLAGS=-ldflags "-w -s"

# Define UPX command (modify if needed)
UPX=upx --brute

# Default target: Build both binaries
all: $(CONTROLLER_BIN) $(RUNNER_BIN)

# Build the controller binary
$(CONTROLLER_BIN): $(shell find ./cmd/controller -type f) $(shell find ./internal -type f)
	@echo "Building $(CONTROLLER_BIN)..."
	go build $(BUILD_FLAGS) -o $(CONTROLLER_BIN) ./cmd/controller

# Build the runner binary and compress it with UPX
$(RUNNER_BIN): $(shell find ./cmd/runner -type f) $(shell find ./internal -type f)
	@echo "Building $(RUNNER_BIN)..."
	go build $(BUILD_FLAGS) -o $(RUNNER_BIN) ./cmd/runner
	@echo "Compressing $(RUNNER_BIN) with UPX..."
	$(UPX) $(RUNNER_BIN)

# Clean up binaries
clean:
	@echo "Cleaning up binaries..."
	rm -f $(CONTROLLER_BIN) $(RUNNER_BIN) $(BUILD_BIN)

# Rebuild both binaries
rebuild: clean all
	
$(BUILD_BIN): $(shell find ./cmd/build -type f) $(shell find ./internal -type f)
	go build -o $(BUILD_BIN) ./cmd/build

.PHONY: build-playbook
build-playbook: $(BUILD_BIN)
	./$(BUILD_BIN) --playbook playbook.yaml --out dist

.PHONY: all clean rebuild build-playbook
