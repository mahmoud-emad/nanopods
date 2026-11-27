help:
	@echo "Available commands:"
	@echo "  build - Build the project"
	@echo "  clean - Clean the build artifacts"

build:
	@echo "Building the project..."
	go build -o nanopods main.go

clean:
	@echo "Cleaning build artifacts..."
	rm -f nanopods