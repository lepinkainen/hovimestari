# LLM Integration

## LLM Interaction

The application interacts with Google's Gemini LLM through the `internal/llm/gemini.go` module:

1. Prompt templates are stored in `prompts.json` with placeholders for dynamic content
2. The `BuildBriefPrompt` method combines memories, user context, and the prompt template
3. The `Generate` method sends the prompt to the Gemini API and receives the response
4. The response is returned to the user in the specified output format(s)

Prompts include placeholders like `%LANG%` for output language, `%NOTES%` for memories, and `%CONTEXT%` for user context information.

## LLM Providers

The application currently supports one LLM provider:

- **Google Gemini**: Cloud-based LLM service with API key authentication

Support for additional LLM providers (including Ollama, OpenAI, Anthropic Claude, etc.) is planned for future development as outlined in `docs/project-plan.md`.

## Prompt Structure

Prompts are defined in `prompts.json` and include detailed instructions for the LLM:

- **dailyBrief**: Template for generating daily briefs
- **userQuery**: Template for responding to user queries

Each prompt includes placeholders for dynamic content and specific instructions on tone, structure, and content:

- **%CONTEXT%**: Placeholder for context information (date, time, weather, etc.)
- **%NOTES%**: Placeholder for memories/notes
- **%LANG%**: Placeholder for output language
- **%QUERY%**: Placeholder for user queries (in userQuery prompt)
