# Water Quality Importer Implementation Plan

This plan outlines the steps to add a new CLI command `import-water-quality` to Hovimestari. This command will allow for adding water quality data for specific locations to the database as "memories."

## Phase 1: Core Implementation

- [ ] **Create a new command file for the water quality importer.**

  - Create file: `cmd/hovimestari/commands/import_water_quality.go`

- [ ] **Define the `import-water-quality` CLI command.**

  - Use the Cobra framework within the new file to define the command.
  - The command should be named `import-water-quality`.
  - It should accept two required string flags:
    - `--location`: The name of the measurement location (e.g., "Hietaniemi").
    - `--quality`: The water quality status (e.g., "Excellent", "Good", "Poor").

- [ ] **Implement the command's execution logic.**

  - The command's `Run` function should perform the following actions:
    - Load the application configuration and initialize the database store (`store.Store`).
    - Create a new `store.Memory` object.
    - Populate the `Memory` object:
      - `Content`: Format a string like: "Water quality at {location} is {quality}."
      - `Source`: Create a unique source identifier, e.g., `waterquality:{location}`.
      - `RelevanceDate`: Set to the current timestamp.
    - Use the existing `store.AddMemory()` function to save the new memory to the database.
    - Print a confirmation message to the user (e.g., "Successfully added water quality memory for {location}.").

- [ ] **Register the new command in the main application.**
  - Update `cmd/hovimestari/main.go` to add the new `import-water-quality` command to the root command.

## Phase 2: Configuration (Future-proofing)

_While the initial implementation will use CLI flags, setting up configuration will make it easier to add automated fetching from an API later._

- [ ] **Update the configuration structure.**

  - Modify `internal/config/viper.go` to add a new section for water quality.
  - Add a `WaterQualityLocations` slice to the `Config` struct. Each item in the slice could be a struct with a `Name` field.

  ```go
  type WaterQualityLocation struct {
      Name string `mapstructure:"name"`
  }
  ...
  type Config struct {
      // ... existing fields
      WaterQualityLocations []WaterQualityLocation `mapstructure:"water_quality_locations"`
  }
  ```

- [ ] **Update the example configuration file.**
  - Add a sample `water_quality_locations` array to `config.example.json`.
  ```json
  "water_quality_locations": [
    { "name": "Hietaniemi" },
    { "name": "Piraeus" }
  ],
  ```
- [ ] **(Optional) Add validation logic.**
  - Modify the `import-water-quality` command to check if the provided `--location` flag value exists in the configuration. This ensures data is only added for pre-approved locations.

## Phase 3: Documentation

- [ ] **Update the CLI documentation.**

  - Edit `docs/07_cli.md` to add the `import-water-quality` command, its purpose, and its flags.

- [ ] **Update the configuration documentation.**

  - Edit `docs/04_configuration.md` to explain the new `water_quality_locations` section in the `config.json` file.

- [ ] **Update the main README.**
  - Briefly mention the new capability in the `README.md` file if appropriate.
