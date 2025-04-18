.PHONY: build clean run import-calendar generate-brief

# Build the application
build:
	go build -o hovimestari ./cmd/hovimestari

# Clean build artifacts
clean:
	rm -f hovimestari

# Run the application
run: build
	./hovimestari

# Import calendar events
import-calendar: build
	./hovimestari import-calendar

# Generate a daily brief
generate-brief: build
	./hovimestari generate-brief

# Initialize the configuration (requires GEMINI_API_KEY and WEBCAL_URL)
init-config: build
	./hovimestari init-config --gemini-api-key="$(GEMINI_API_KEY)" --webcal-url="$(WEBCAL_URL)"

# Add a memory (requires CONTENT, optional RELEVANCE_DATE and SOURCE)
add-memory: build
	./hovimestari add-memory --content="$(CONTENT)" $(if $(RELEVANCE_DATE),--relevance-date="$(RELEVANCE_DATE)") $(if $(SOURCE),--source="$(SOURCE)")

# Install dependencies
deps:
	go mod tidy

# Run tests
test:
	go test ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  clean           - Clean build artifacts"
	@echo "  run             - Run the application"
	@echo "  import-calendar - Import calendar events"
	@echo "  generate-brief  - Generate a daily brief"
	@echo "  init-config     - Initialize the configuration (requires GEMINI_API_KEY and WEBCAL_URL)"
	@echo "  add-memory      - Add a memory (requires CONTENT, optional RELEVANCE_DATE and SOURCE)"
	@echo "  deps            - Install dependencies"
	@echo "  test            - Run tests"
	@echo "  help            - Show this help message"
