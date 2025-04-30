# Project Overview

Hovimestari ("Butler" in Finnish) is a personal AI butler assistant that generates daily briefs in Finnish. It collects "memories" from various sources (calendar events, weather forecasts, manual notes) and uses Google's Gemini LLM to generate personalized daily summaries.

## Technology Stack

- **Language:** Go (Golang)
- **Database:** SQLite (using `modernc.org/sqlite`)
- **LLM:** Google Gemini (using `github.com/google/generative-ai-go`)
- **Calendar Parsing:** iCalendar (using `github.com/apognu/gocal`)
- **Weather API:** MET Norway Locationforecast API
- **CLI Framework:** Cobra (`github.com/spf13/cobra`)
- **Build System:** Task (`github.com/go-task/task`) with Taskfile.yml
