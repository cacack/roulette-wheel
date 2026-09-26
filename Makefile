.PHONY: build clean run test test-short

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

# Run all tests, including the statistical fairness simulation
test:
	go test ./...

# Run tests without the statistical fairness simulation
test-short:
	go test -short ./...
