# Hovimestari Project Context

This document provides detailed technical information about the Hovimestari project structure, architecture, and implementation details. It is designed to help LLM agents and developers understand the project quickly.

## Project Overview

Hovimestari ("Butler" in Finnish) is a personal AI butler assistant that generates daily briefs in Finnish. It collects "memories" from various sources (calendar events, weather forecasts, manual notes) and uses Google's Gemini LLM to generate personalized daily summaries.

## Technology Stack

- **Language:** Go (Golang)
- **Database:** SQLite (using `modernc.org/sqlite`)
- **LLM:** Google Gemini (using `github.com/google/generative-ai-go`)
- **Calendar Parsing:** iCalendar (using `github.com/apognu/gocal`)
- **Weather API:** MET Norway Locationforecast API
- **CLI Framework:** Cobra (`github.com/spf13/cobra`)
- **Build System:** Task (`github.com/go-task/task`) with Taskfile.yml

## Directory Structure

```
hovimestari/
├── build/                # Build artifacts directory
├── cmd/
│   └── hovimestari/
│       └── main.go       # Main application entry point and CLI commands
├── internal/
│   ├── brief/
│   │   └── brief.go      # Handles daily brief generation
│   ├── config/
│   │   └── config.go     # Configuration management
│   ├── importer/
│   │   ├── calendar/
│   │   │   └── calendar.go   # Calendar event importing
│   │   └── weather/
│   │       └── weather.go    # Weather forecast importing
│   ├── llm/
│   │   └── gemini.go     # Google Gemini API client
│   ├── output/
│   │   └── output.go     # Output handling (CLI, Discord, Telegram)
│   ├── store/
│   │   └── store.go      # SQLite database operations
│   └── weather/
│       └── metno.go      # MET Norway API client
├── config.example.json   # Example configuration file
├── prompts.json          # LLM prompt templates
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Taskfile.yml          # Task runner configuration
└── README.md             # User documentation
```

## Key Files

- **cmd/hovimestari/main.go**: Contains the CLI command definitions using Cobra. Defines commands for importing calendar events, importing weather forecasts, generating briefs, adding memories manually, initializing configuration, listing available Gemini models, and showing brief context. Integrates with the output package to send briefs to multiple destinations (CLI, Discord, Telegram).

- **internal/brief/brief.go**: Handles the generation of daily briefs by combining memories from the database with context information (date, time, weather, birthdays, etc.) and sending them to the LLM.

- **internal/config/config.go**: Manages loading and saving application configuration from `config.json`. Defines the configuration structure including database path, API keys, location information, calendars, family members, and output settings.

- **internal/importer/calendar/calendar.go**: Fetches and parses calendar events from WebCal URLs and stores them as memories in the database.

- **internal/importer/weather/weather.go**: Imports weather forecasts from the MET Norway API and stores them as memories.

- **internal/llm/gemini.go**: Provides the client for interacting with the Google Gemini API, including methods for generating briefs and responses to user queries.

- **internal/output/output.go**: Implements different output methods for sending briefs to various destinations, including CLI (terminal), Discord (via webhooks), and Telegram (via bot API).

- **internal/store/store.go**: Manages the SQLite database connection and operations for adding and querying memories.

- **internal/weather/metno.go**: Fetches weather forecasts from the MET Norway Locationforecast API.

- **prompts.json**: Contains the prompt templates used for generating briefs and responses to user queries.

- **config.example.json**: An example configuration file showing the required structure and fields, including output configuration options.

- **Taskfile.yml**: Task runner configuration for building, testing, and running the application. Defines tasks for common operations like building for different platforms, running tests, and executing application commands.

## Core Concepts

### Memories

Memories are the fundamental data units in Hovimestari. Each memory represents a piece of information stored in the SQLite database with the following attributes:

- **Content**: The actual information (e.g., "Calendar Event: Meeting with John from 2025-04-20 14:00 to 15:00")
- **CreatedAt**: When the memory was added to the database
- **RelevanceDate**: When the memory is relevant (e.g., the date of a calendar event)
- **Source**: Where the memory came from (e.g., "calendar:work", "weather:helsinki", "manual")
- **UID**: Optional unique identifier (used for calendar events to prevent duplicates)

### Importers

Importers fetch data from external sources and store them as memories in the database:

- **Calendar Importer**: Fetches events from WebCal/iCalendar URLs
- **Weather Importer**: Fetches weather forecasts from the MET Norway API

### Briefs

Briefs are daily summaries generated by the LLM based on relevant memories. They include:

- Today's date and weather
- Ongoing events
- Today's calendar events
- Upcoming days' events and weather
- Special notifications (birthdays, high UV index, etc.)

Briefs are generated in Finnish with a formal, butler-like tone.

### Configuration

The application is configured through a `config.json` file with the following key sections:

- **Database**: Path to the SQLite database file
- **LLM**: Google Gemini API key, model name, and output language
- **Location**: Name, coordinates, and timezone for weather forecasts
- **Calendars**: List of calendars to import events from
- **Family**: List of family members with optional birthdays and Telegram IDs
- **Output**: Configuration for different output methods (CLI, Discord, Telegram)

## Data Flow

```mermaid
graph TD
    subgraph Importers
        Calendar[Calendar Importer]
        WeatherImporter[Weather Importer]
        Manual[Manual Add Command]
    end

    subgraph Core
        Store[(SQLite DB\nmemories.db)]
        LLM[Gemini LLM]
        Config[config.json]
        Prompts[prompts.json]
        Output[Output Module]
    end

    subgraph Commands
        ImportCmd[Import Commands]
        BriefCmd[Generate Brief Command]
        AddCmd[Add Memory Command]
    end

    subgraph Outputs
        CLI[CLI Output]
        Discord[Discord Webhooks]
        Telegram[Telegram Bots]
    end

    User(User) --> AddCmd
    User --> ImportCmd
    User --> BriefCmd

    ImportCmd --> Calendar
    ImportCmd --> WeatherImporter
    AddCmd --> Manual

    Calendar -- Stores --> Store
    WeatherImporter -- Stores --> Store
    Manual -- Stores --> Store

    BriefCmd -- Reads --> Store
    BriefCmd -- Uses --> LLM
    BriefCmd -- Reads --> Config
    LLM -- Reads --> Prompts

    BriefCmd -- Sends to --> Output
    Output -- Outputs to --> CLI
    Output -- Outputs to --> Discord
    Output -- Outputs to --> Telegram

    style User fill:#f9f,stroke:#333,stroke-width:2px
    style CLI fill:#ccf,stroke:#333,stroke-width:2px
    style Discord fill:#ccf,stroke:#333,stroke-width:2px
    style Telegram fill:#ccf,stroke:#333,stroke-width:2px
```

## Database Schema

The SQLite database (`memories.db`) has a single table called `memories` with the following schema:

```sql
CREATE TABLE memories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    relevance_date TIMESTAMP,
    source TEXT NOT NULL,
    uid TEXT
);
```

Indexes are created on `relevance_date`, `source`, and the combination of `source` and `uid` to optimize queries.

## LLM Interaction

The application interacts with Google's Gemini LLM through the `internal/llm/gemini.go` module:

1. Prompt templates are stored in `prompts.json` with placeholders for dynamic content
2. The `BuildBriefPrompt` method combines memories, user context, and the prompt template
3. The `Generate` method sends the prompt to the Gemini API and receives the response
4. The response is returned to the user in the specified output format(s)

Prompts include placeholders like `%LANG%` for output language, `%NOTES%` for memories, and `%CONTEXT%` for user context information.

## CLI Commands

The application provides several CLI commands through the Cobra framework:

- **import-calendar**: Import events from configured WebCal URLs
- **import-weather**: Import weather forecasts for the configured location
- **generate-brief**: Generate a daily brief based on stored memories
  - `--days-ahead`: Number of days ahead to include in the brief (default: 2)
- **add-memory**: Add a memory manually
- **init-config**: Initialize the configuration file
  - `--output-format`: Output format (cli, telegram)
- **list-models**: List available Gemini models
- **show-brief-context**: Show the context that would be sent to the LLM

## Configuration File Example

```json
{
  "db_path": "memories.db",
  "gemini_api_key": "YOUR_GEMINI_API_KEY",
  "gemini_model": "gemini-2.0-flash",
  "outputLanguage": "Finnish",
  "promptFilePath": "prompts.json",
  "location_name": "Helsinki",
  "latitude": 60.1699,
  "longitude": 24.9384,
  "timezone": "Europe/Helsinki",
  "calendars": [
    {
      "name": "Family Calendar",
      "url": "webcal://p12-caldav.icloud.com/published/2/MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI"
    },
    {
      "name": "Work Calendar",
      "url": "webcal://example.com/work-calendar.ics"
    }
  ],
  "family": [
    {
      "name": "Matti",
      "birthday": "1980-05-15",
      "telegram_id": "matti_v"
    },
    {
      "name": "Maija",
      "birthday": "1982-11-22"
    },
    {
      "name": "Pekka",
      "birthday": "2010-01-30"
    }
  ],
  "output_format": "cli",
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

## Prompt Structure

Prompts are defined in `prompts.json` and include detailed instructions for the LLM:

- **dailyBrief**: Template for generating daily briefs
- **userQuery**: Template for responding to user queries

Each prompt includes placeholders for dynamic content and specific instructions on tone, structure, and content.
