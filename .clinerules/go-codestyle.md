### Code Style

- Follow standard Go code style and conventions (gofmt).
- Use meaningful variable and function names.
- Add comments for exported functions and types.
- Keep functions small and focused on a single responsibility.

### Error Handling

- Use the `fmt.Errorf("failed to X: %w", err)` pattern for error wrapping.
- Always check errors and provide context.
- Avoid panics in production code.

### Dependencies

- Prefer standard library solutions when possible.
- Minimize external dependencies, unless an external dependency makes the code more cleaner or efficient.
- Pin dependency versions in go.mod.

### Testing

- Write unit tests for core functionality.
- Use table-driven tests where appropriate.
- Mock external dependencies for testing.

### Development

- Run unit tests after a task is complete and confirm that they pass
