BIN_DIR := bin

.PHONY: all build clean logger node test test-scenario1 test-scenario2

all: build

build: node

node:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/node ./cmd/node
	cp $(BIN_DIR)/node mp1_node

clean:
	rm -rf $(BIN_DIR)
