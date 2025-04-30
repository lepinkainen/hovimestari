# Configuration

The application is configured through a `config.json` file with the following key sections:

- **Database**: Path to the SQLite database file
- **LLM**: Provider (currently only Gemini), API key, model name, and output language
- **Location**: Name, coordinates, and timezone for weather forecasts
- **Calendars**: List of calendars to import events from
- **Family**: List of family members with optional birthdays and Telegram IDs
- **Output**: Configuration for different output methods (CLI, Discord, Telegram)

The configuration system uses Spf13/Viper for robust configuration management, supporting:

- Multiple configuration sources (file, environment variables)
- XDG Base Directory Specification for standard file locations
- Comprehensive validation of configuration values

## Configuration File Example

### Standard Configuration (Gemini)

```json
{
  "db_path": "memories.db",
  "gemini_api_key": "YOUR_GEMINI_API_KEY",
  "gemini_model": "gemini-2.0-flash",
  "outputLanguage": "Finnish",
  "promptFilePath": "prompts.json",
  "days_ahead": 2,
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

### Planned Ollama Configuration (Not Yet Implemented)

```json
{
  "db_path": "memories.db",
  "llm_provider": "ollama",
  "ollama_url": "http://localhost:11434",
  "ollama_model": "llama3",
  "outputLanguage": "Finnish",
  "promptFilePath": "prompts.json",
  "days_ahead": 2,
  "location_name": "Helsinki",
  "latitude": 60.1699,
  "longitude": 24.9384,
  "timezone": "Europe/Helsinki",
  "calendars": [
    {
      "name": "Family Calendar",
      "url": "webcal://example.com/family-calendar.ics"
    }
  ],
  "family": [
    {
      "name": "Matti",
      "birthday": "1980-05-15"
    }
  ],
  "outputs": {
    "enable_cli": true
  }
}
```

Note: This configuration is for future reference once Ollama support is implemented as outlined in `docs/project-plan.md`.
