# NOTE

This project has been 99.5% ✨Vibe Coded✨ with [Cline](https://cline.bot), using `Gemini 2.5-pro-preview-03-25` for Planning mode and `Claude-3-7-sonnet-20250219` for Act mode

The project started by me giving Geoffrey Litt's page on their AI butler assistant "Stevens" (linked below) to Cline and off we went. At this point (v1.0-ish) the context sizes have grown so much that every change costs more than I'm willing to spend. Trying to keep Claude 3.7 on track with prompts is also becoming more and more difficult, so I'll most likely move to coding the ye olde way from here on out.

-- END OF HUMAN GENERATED CONTENT --

# Hovimestari - A Personal AI Butler Assistant

Hovimestari is a personal AI butler assistant inspired by the article ["How I Made a Useful AI Assistant with One SQLite Table and a Handful of Cron Jobs"](https://www.geoffreylitt.com/2025/04/12/how-i-made-a-useful-ai-assistant-with-one-sqlite-table-and-a-handful-of-cron-jobs) by Geoffrey Litt.

The name "Hovimestari" means "Butler" in Finnish.

## Features

- **Calendar Integration**: Import events from multiple iCloud/WebCal calendars
- **Weather Forecasts**: Automatically fetches weather data from MET Norway API
- **Family Information**: Keeps track of family members and their birthdays
- **Daily Briefs**: Generates personalized daily briefs in Finnish with a formal butler tone
- **Memory Storage**: Stores all information in a simple SQLite database
- **Multiple Output Options**: Send briefs to CLI, Discord webhooks, and Telegram bots

## Overview

Hovimestari is a simple yet powerful personal assistant that:

1. Stores "memories" in a single SQLite database table
2. Imports calendar events from an iCloud WebCal URL
3. Generates daily briefs using Google's Gemini AI
4. Allows manual addition of memories
5. Uses a formal, butler-like tone in its communications (in Finnish)
6. Supports multiple output destinations (CLI, Discord, Telegram)

## Architecture

The application follows a clean, modular architecture with a focus on simplicity and extensibility. For detailed technical information about the project structure, components, and implementation details, please refer to [PROJECT_CONTEXT.md](PROJECT_CONTEXT.md).

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
task deps
```

3. Build the application:

```bash
task build
```

### Cross-Compilation for Linux

The application uses a pure Go SQLite implementation via `modernc.org/sqlite`, which doesn't require CGO. This makes cross-compilation simple and straightforward:

```bash
# Build for Linux
task build-linux
```

No additional dependencies or cross-compilers are needed, making it easy to build for any platform from any platform.

4. Initialize the configuration:

```bash
# Using environment variables
GEMINI_API_KEY="YOUR_GEMINI_API_KEY" WEBCAL_URL="YOUR_WEBCAL_URL" task init-config

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
# Using Task
task import-calendar

# Or directly with the CLI
./hovimestari import-calendar
```

Hovimestari supports different update modes for calendar imports, allowing you to choose between "smart" updates for static calendars and "full_refresh" for dynamic ones. See [docs/calendar-import.md](docs/calendar-import.md) for details.

#### Generate a Daily Brief

```bash
# Using Task
task generate-brief

# Or directly with the CLI
./hovimestari generate-brief
```

#### Add a Memory Manually

```bash
# Using Task with environment variables
task add-memory CONTENT="Remember to buy milk" RELEVANCE_DATE="2025-04-20" SOURCE="manual"

# Or directly with the CLI
./hovimestari add-memory --content="Remember to buy milk" --relevance-date="2025-04-20" --source="manual"
```

#### Available Task Commands

Run `task --list` to see all available tasks:

```
task: Available tasks for this project:
* add-memory:       Add a memory (reads CONTENT, RELEVANCE_DATE, SOURCE from env/args)
* build:            Build the Go application for the current OS/ARCH
* build-linux:      Build the Go application for Linux AMD64
* clean:            Clean build artifacts
* default:          Build the application (default task)
* deps:             Tidy Go module dependencies
* generate-brief:   Generate a daily brief
* import-calendar:  Import calendar events
* import-weather:   Import weather forecasts
* init-config:      Initialize the configuration (reads GEMINI_API_KEY, WEBCAL_URL from env)
* lint:             Run Go linters (requires golangci-lint)
* run:              Build and run the application
* test:             Run Go tests
* upgrade-deps:     Upgrade all dependencies to their latest versions
```

## Configuration

### Configuration Files

Hovimestari uses the following configuration files:

- **config.json**: Main configuration file with API keys, location settings, etc.
- **prompts.json**: Contains the prompts used for generating briefs and responses
- **memories.db**: SQLite database storing all memories and calendar events

### Configuration File Locations

Hovimestari follows the XDG Base Directory Specification for configuration files. It looks for files in the following order:

1. The path specified with the `--config` flag (for `config.json` only)
2. `$XDG_CONFIG_HOME/hovimestari/` (usually `~/.config/hovimestari/`)
3. The directory containing the executable

This makes it easy to run Hovimestari from cron jobs without having to specify the full path to each file.

You can specify a different path for the main configuration file using the `--config` flag:

```bash
./hovimestari generate-brief --config=/path/to/your/config.json
```

You can also control the logging level using the `--log-level` flag:

```bash
./hovimestari generate-brief --log-level=debug
```

Valid log levels are: debug, info, warn, error (default is "debug" for the flag, "info" for the config file).

```json
{
  "db_path": "memories.db",
  "gemini_api_key": "YOUR_GEMINI_API_KEY",
  "location_name": "Helsinki",
  "latitude": 60.1699,
  "longitude": 24.9384,
  "timezone": "Europe/Helsinki",
  "calendars": [
    {
      "name": "Family Calendar",
      "url": "webcal://example.com/calendar.ics",
      "update_mode": "full_refresh"
    },
    {
      "name": "Holidays",
      "url": "https://calendars.icloud.com/holidays/fi_fi.ics/",
      "update_mode": "smart"
    }
  ],
  "family": [
    {
      "name": "Matti",
      "birthday": "1980-05-15"
    }
  ],
  "outputs": {
    "enable_cli": true,
    "discord_webhook_urls": [
      "https://discord.com/api/webhooks/your-webhook-id/your-webhook-token"
    ],
    "telegram_bots": [
      {
        "bot_token": "YOUR_TELEGRAM_BOT_TOKEN",
        "chat_id": "YOUR_TELEGRAM_CHAT_ID"
      }
    ]
  }
}
```

### Configuration Fields

- **db_path**: Path to the SQLite database file (defaults to `$XDG_CONFIG_HOME/hovimestari/memories.db` if not specified)
- **log_level**: Logging level (debug, info, warn, error) - defaults to "info"
- **gemini_api_key**: Your Google Gemini API key
- **gemini_model**: Gemini model to use (e.g., "gemini-2.0-flash") - defaults to "gemini-2.0-flash"
- **outputLanguage**: Language for LLM responses (e.g., "Finnish", "English") - defaults to "Finnish"
- **days_ahead**: Number of days ahead to include in the brief - defaults to 2
- **location_name**: Name of your location (e.g., "Helsinki")
- **latitude** and **longitude**: Geographic coordinates for weather forecasts
- **timezone**: Your timezone in IANA format (e.g., "Europe/Helsinki")
- **calendars**: List of calendars to import events from
  - **name**: Name of the calendar
  - **url**: WebCal URL for the calendar
  - **update_mode**: Update strategy for the calendar ("smart" or "full_refresh", defaults to "full_refresh")
- **family**: List of family members with optional information
  - **name**: Name of the family member
  - **birthday**: Optional birthday in YYYY-MM-DD format
  - **telegram_id**: Optional Telegram ID for the family member
- **output_format**: Legacy field for output format (cli, telegram, etc.) - use **outputs** instead
- **outputs**: Configuration for multiple output methods:
  - **enable_cli**: Whether to output to the command line
  - **discord_webhook_urls**: List of Discord webhook URLs to send briefs to
  - **telegram_bots**: List of Telegram bot configurations, each with a bot token and chat ID

## Testing

The project includes unit tests for deterministic functions that don't rely on external dependencies like network connections, databases, or LLM services.

### Running Tests

To run all tests:

```bash
task test
```

Or directly with Go:

```bash
go test ./...
```

### Test Coverage

The following deterministic functions are covered by unit tests:

#### Calendar Importer (`internal/importer/calendar`)

- **URL Conversion in `NewImporter`**: Tests the conversion of `webcal://` URLs to `https://`
- **`formatEvent`**: Tests the formatting of calendar events into memory strings
- **`filterValidEvents`**: Tests the filtering of iCalendar data to remove events without DTSTAMP

#### Weather Package (`internal/weather`)

- **`FormatDailyForecast`**: Tests the formatting of weather forecast data into human-readable strings

#### Weather Importer (`internal/importer/weather`)

- **`NewImporter`**: Tests the creation of a new weather importer with correct parameters
- **`SourcePrefix`**: Tests that the source prefix constant is set correctly

### Adding New Tests

When adding new functionality to the project, consider writing unit tests for deterministic functions that don't rely on external dependencies. This helps ensure that the core logic remains correct as the codebase evolves.

## Future Enhancements

- Additional importers (email, news, etc.)
- Web interface for viewing and managing memories
- Scheduled execution via cron jobs
- Expanded test coverage

## License

MIT

## Acknowledgements

This project was inspired by Geoffrey Litt's article on building a personal AI assistant.
