.PHONY: build clean run

BINARY_NAME=roulette-wheel
DIST_DIR=dist

# Build for current platform/architecture
build: clean
	@mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(DIST_DIR)

run:
	go run .
