# Hovimestari - A Personal AI Butler Assistant

Hovimestari is a personal AI butler assistant inspired by the article ["How I Made a Useful AI Assistant with One SQLite Table and a Handful of Cron Jobs"](https://www.geoffreylitt.com/2025/04/12/how-i-made-a-useful-ai-assistant-with-one-sqlite-table-and-a-handful-of-cron-jobs) by Geoffrey Litt.

The name "Hovimestari" means "Butler" in Finnish, and this assistant generates its briefs in Finnish.

## Overview

Hovimestari is a simple yet powerful personal assistant that:

1. Stores "memories" in a single SQLite database table
2. Imports calendar events from an iCloud WebCal URL
3. Generates daily briefs using Google's Gemini AI
4. Allows manual addition of memories
5. Uses a formal, butler-like tone in its communications (in Finnish)

## Architecture

The application follows a clean, modular architecture:

- **Database**: A single SQLite table stores all memories
- **Importers**: Components that fetch data from external sources and store them as memories
- **LLM Client**: Interfaces with Google's Gemini API to generate human-like text
- **Brief Generator**: Creates daily briefs based on relevant memories
- **CLI**: Command-line interface for interacting with the system

## Getting Started

### Prerequisites

- Go 1.21 or later
- Google Gemini API key
- iCloud WebCal URL

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/hovimestari.git
cd hovimestari
```

2. Install dependencies:

```bash
make deps
```

3. Build the application:

```bash
make build
```

4. Initialize the configuration:

```bash
# Using environment variables
GEMINI_API_KEY="YOUR_GEMINI_API_KEY" WEBCAL_URL="YOUR_WEBCAL_URL" make init-config

# Or directly with the CLI
./hovimestari init-config --gemini-api-key="YOUR_GEMINI_API_KEY" --webcal-url="YOUR_WEBCAL_URL"
```

Alternatively, you can copy the example configuration file and edit it:

```bash
cp config.example.json config.json
# Edit config.json with your favorite editor
```

### Usage

You can use the CLI directly or use the provided Makefile targets.

#### Import Calendar Events

```bash
# Using make
make import-calendar

# Or directly with the CLI
./hovimestari import-calendar
```

#### Generate a Daily Brief

```bash
# Using make
make generate-brief

# Or directly with the CLI
./hovimestari generate-brief
```

#### Add a Memory Manually

```bash
# Using make with environment variables
CONTENT="Remember to buy milk" RELEVANCE_DATE="2025-04-20" SOURCE="manual" make add-memory

# Or directly with the CLI
./hovimestari add-memory --content="Remember to buy milk" --relevance-date="2025-04-20" --source="manual"
```

#### Available Make Targets

Run `make help` to see all available targets:

```
Available targets:
  build           - Build the application
  clean           - Clean build artifacts
  run             - Run the application
  import-calendar - Import calendar events
  generate-brief  - Generate a daily brief
  init-config     - Initialize the configuration (requires GEMINI_API_KEY and WEBCAL_URL)
  add-memory      - Add a memory (requires CONTENT, optional RELEVANCE_DATE and SOURCE)
  deps            - Install dependencies
  test            - Run tests
  help            - Show this help message
```

## Configuration

The configuration is stored in `config.json` by default. You can specify a different path using the `--config` flag.

```json
{
  "db_path": "memories.db",
  "gemini_api_key": "YOUR_GEMINI_API_KEY",
  "webcal_url": "YOUR_WEBCAL_URL",
  "output_format": "cli"
}
```

## Future Enhancements

- Telegram integration for sending daily briefs and receiving queries
- Additional importers (weather, email, etc.)
- Web interface for viewing and managing memories
- Scheduled execution via cron jobs

## License

MIT

## Acknowledgements

This project was inspired by Geoffrey Litt's article on building a personal AI assistant.
