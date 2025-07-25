# Build the plumbops binary
BINARY=plumbops
BUILD_FLAGS=-ldflags "-w -s"
UPX=upx --brute

all: $(BINARY)

$(BINARY): $(shell find ./cmd/plumbops -type f) $(shell find ./internal -type f)
	@echo "Building $(BINARY)..."
	go build $(BUILD_FLAGS) -o $(BINARY) ./cmd/plumbops

clean:
	@echo "Cleaning up binaries..."
	rm -f $(BINARY)

rebuild: clean all

.PHONY: all clean rebuild
