# GitHub Issues Report - Hovimestari

**Generated:** 2025-12-30
**Repository:** lepinkainen/hovimestari

## Summary

- **Total Open Issues:** 29
- **Critical Issues:** 1
- **High Priority Issues:** 2
- **Medium Priority Issues:** 8
- **Low Priority Issues:** 3
- **Feature Requests:** 8
- **Recently Closed:** Issue #28 (godoc comments added)

---

## Critical Priority Issues

### #4 - CRITICAL: Security vulnerability - Telegram bot token exposed in debug logs

**Labels:** Security, High-Priority
**Status:** OPEN
**File:** `internal/output/telegram.go:95`

**Problem:**
The Telegram bot token is being logged in debug messages, creating a critical security vulnerability where sensitive credentials could be exposed in log files, console output, or log aggregation systems.

**Current Code:**

```go
slog.Debug("Sending HTTP request to Telegram API", "url", url)
// url contains: https://api.telegram.org/bot<TOKEN>/sendMessage
```

**Required Fix:**

```go
slog.Debug("Sending HTTP request to Telegram API", "host", "api.telegram.org", "chat_id", o.ChatID)
```

**Impact:**

- Credential exposure in logs
- Potential unauthorized bot access
- Compliance violations
- Should be fixed **immediately** before any deployment

---

## High Priority Issues

### #5 - Missing HTTP timeouts causing potential hanging requests

**Labels:** Performance, Reliability
**Status:** OPEN
**Files:** `internal/weather/metno.go`, `internal/importer/calendar/calendar.go`, `internal/output/telegram.go`, `internal/output/discord.go`

**Problem:**
HTTP clients throughout the codebase use `&http.Client{}` without timeout configurations, which can cause the application to hang indefinitely when external services are slow or unresponsive.

**Affected Locations:**

- `internal/weather/metno.go:81, 177, 316`
- `internal/output/telegram.go:96`
- `internal/output/discord.go:58`

**Required Fix:**

```go
client := &http.Client{
    Timeout: 30 * time.Second,
}
```

**Impact:**

- Application can hang indefinitely
- Poor user experience
- Potential resource exhaustion

---

### #19 - Add tests for internal/config/viper.go

**Labels:** High-Priority, Testing, Code-Health
**Status:** OPEN
**File:** `internal/config/viper.go` (384 lines, no tests)

**Problem:**
Configuration loading, validation, and file management is critical infrastructure but has no tests. Configuration errors can break the entire application.

**Missing Test Coverage:**

- `InitViper()` - Config initialization
- `GetConfig()` - Config retrieval and validation
- `LoadPrompts()` - Prompt file loading
- `FindConfigFile()` - Config file resolution (XDG paths)
- Validation of required fields
- Default value handling

---

## Medium Priority Issues

### #6 - Insufficient test coverage for core business logic

**Labels:** Testing, Code-Health
**Status:** OPEN

**Problem:**
Overall insufficient test coverage across the codebase, particularly for core business logic.

---

### #17 - Add comprehensive tests for internal/store/store.go

**Labels:** Testing, Code-Health
**Status:** OPEN
**File:** `internal/store/store.go` (434 lines, no tests)

**Missing Test Coverage:**

- Memory operations (Add, Get, Exists, GetBySource)
- Calendar event operations (Add, Update, Delete, Get)
- Database initialization
- Error handling

---

### #18 - Add comprehensive tests for internal/brief/brief.go

**Labels:** Testing, Code-Health
**Status:** OPEN
**File:** `internal/brief/brief.go` (350 lines, no tests)

**Missing Test Coverage:**

- `BuildBriefContext()` - Context aggregation
- `GenerateDailyBrief()` - Main brief generation
- `findBirthdaysToday()` - Birthday detection
- `getOngoingCalendarEvents()` - Ongoing event detection
- Calendar event formatting

---

### #20 - Refactor: Split internal/store/store.go into smaller, focused files

**Labels:** Medium-Priority, Refactoring, Code-Health
**Status:** OPEN
**File:** `internal/store/store.go` (434 lines)

**Proposed Structure:**

1. `store.go` (core, ~50 lines) - Store struct, constructor, initialization
2. `memory_ops.go` (~150 lines) - All memory-related operations
3. `calendar_ops.go` (~200 lines) - All calendar event operations

---

### #21 - Refactor: Split internal/weather/metno.go into smaller, focused files

**Labels:** Medium-Priority, Refactoring, Code-Health
**Status:** OPEN
**File:** `internal/weather/metno.go` (442 lines)

**Proposed Structure:**

1. `types.go` (~80 lines) - Forecast types and constants
2. `client.go` (~300 lines) - API client code
3. `formatters.go` (~60 lines) - Formatting logic

---

### #22 - Refactor: Split internal/config/viper.go into smaller, focused files

**Labels:** Medium-Priority, Config, Refactoring, Code-Health
**Status:** OPEN
**File:** `internal/config/viper.go` (384 lines)

**Proposed Structure:**

1. `types.go` (~100 lines) - Config structs
2. `loader.go` (~200 lines) - Viper initialization
3. `prompts.go` (~80 lines) - Prompt loading

---

### #23 - Refactor: Split internal/brief/brief.go into smaller, focused files

**Labels:** Medium-Priority, Refactoring, Code-Health
**Status:** OPEN
**File:** `internal/brief/brief.go` (350 lines)

**Proposed Structure:**

1. `generator.go` (~100 lines) - Core public API
2. `context_builder.go` (~150 lines) - Aggregation logic
3. `formatters.go` (~100 lines) - Formatting functions

---

### #24 - Add tests for internal/xdg/xdg.go

**Labels:** Medium-Priority, Testing, Code-Health
**Status:** OPEN
**File:** `internal/xdg/xdg.go` (109 lines, no tests)

**Missing Test Coverage:**

- `GetConfigDir()` - XDG directory resolution
- `GetExecutableDir()` - Executable directory discovery
- `GetConfigPath()` - Full config path construction
- `FindConfigFile()` - Config file discovery logic

---

### #25 - Add tests for internal/logging/handler.go

**Labels:** Medium-Priority, Testing, Code-Health
**Status:** OPEN
**File:** `internal/logging/handler.go` (105 lines, no tests)

**Missing Test Coverage:**

- `NewHumanReadableHandler()` - Handler initialization
- `Handle()` - Log message formatting
- Level-based formatting (ERROR, WARN, INFO, DEBUG)
- Attribute formatting

---

### #26 - Add comprehensive tests for output packages

**Labels:** Medium-Priority, Testing, Code-Health
**Status:** OPEN
**Files:** `internal/output/cli.go`, `discord.go`, `telegram.go`

**Missing Test Coverage:**

- CLI output formatting
- Discord outputter (mock HTTP calls)
- Telegram outputter (mock HTTP calls)
- Error handling for network failures
- Message formatting for each outputter

**Note:** Basic tests exist in `output_test.go` but need expansion.

---

### #27 - Add package-level documentation (README.md files)

**Labels:** Medium-Priority, Documentation, Code-Health
**Status:** OPEN

**Problem:**
Individual packages lack README files explaining their purpose, design decisions, and usage patterns.

**High-Value Packages Needing Documentation:**

1. `internal/store/` - Two-table design rationale
2. `internal/brief/` - Brief generation workflow
3. `internal/importer/` - Importer pattern and conventions
4. `internal/output/` - Multi-destination output system
5. `internal/config/` - Config resolution and XDG compliance
6. `internal/llm/` - Gemini integration
7. `internal/weather/` - MET Norway API integration

---

## Low Priority Issues

### #29 - Consider abstracting importer pattern with shared interface

**Labels:** Enhancement, Low-Priority, Code-Health
**Status:** OPEN

**Problem:**
Calendar and weather importers follow similar patterns but don't share a common interface.

**Proposed Interface:**

```go
type Importer interface {
    Import(ctx context.Context) error
}
```

**Note:** Current approach works well. Only implement if adding more importers or seeing clear benefits.

---

### #30 - Add tests for command files (lower priority)

**Labels:** Low-Priority, Testing, Code-Health
**Status:** OPEN
**Files:** `cmd/hovimestari/commands/*.go` (8 files)

**Problem:**
Command files lack tests. These are thin wrappers over internal packages.

**Priority Rationale:**
Testing internal packages provides better coverage. Command files are low priority because they contain minimal logic.

---

## Feature Requests

### #9 - feat: Add proactive actionable suggestions to daily briefs

**Status:** OPEN
**Description:** Enhance briefs with actionable suggestions based on context.

---

### #10 - feat: Add personalized focus configuration for briefs

**Status:** OPEN
**Description:** Allow users to configure what types of information to emphasize in briefs.

---

### #11 - feat: Add to-do list integration for task management

**Status:** OPEN
**Description:** Integrate with task management systems to include tasks in briefs.

---

### #12 - feat: Add dynamic day type detection for contextual greetings

**Status:** OPEN
**Description:** Detect weekdays, weekends, holidays, and adjust brief tone accordingly.

---

### #13 - feat: Add Air Quality Index (AQI) integration

**Status:** OPEN
**Description:** Include air quality information in daily briefs.

---

### #14 - feat: Add news headlines integration

**Status:** OPEN
**Description:** Integrate news APIs to include relevant headlines in briefs.

---

### #15 - feat: Add "On This Day" historical events feature

**Status:** OPEN
**Description:** Include historical events that happened on this day.

---

### #16 - feat: Add social media integration (advanced)

**Status:** OPEN
**Description:** Advanced feature to integrate social media updates.

---

## Recently Completed

### #7 - log format wrong when updating calendar ✅

**Status:** CLOSED (2025-08-23)
**Description:** Fixed logging format issue in calendar importer.

### #28 - Add godoc comments to output package types ✅

**Status:** CLOSED (2025-12-30)
**Description:** Added proper godoc comments to all three output package types (CLI, Discord, Telegram).

---

## Recommendations

### Immediate Actions (Critical)

1. **Fix #4 immediately** - Remove bot token from debug logs (security vulnerability)

### Short-term Actions (High Priority)

2. **Fix #5** - Add HTTP timeouts to all HTTP clients
2. **Address #19** - Add tests for config package (critical infrastructure)

### Medium-term Actions

4. **Address #6, #17, #18** - Improve test coverage for store and brief packages
2. **Consider #20-23** - Refactor large files for better maintainability
3. **Add tests (#24-26)** - Expand test coverage for supporting packages

### Long-term Actions

7. **Improve documentation (#27)** - Add package-level README files
2. **Evaluate feature requests (#9-16)** - Prioritize based on user needs

---

## Test Coverage Status

**Files WITH tests:**

- ✅ `internal/weather/metno_test.go`
- ✅ `internal/output/output_test.go` (basic)
- ✅ `internal/importer/schoollunch/schoollunch_test.go`

**Files MISSING tests (prioritized):**

- ❌ `internal/config/viper.go` (384 lines) - **HIGH PRIORITY**
- ❌ `internal/store/store.go` (434 lines) - **HIGH PRIORITY**
- ❌ `internal/brief/brief.go` (350 lines) - **HIGH PRIORITY**
- ❌ `internal/xdg/xdg.go` (109 lines)
- ❌ `internal/logging/handler.go` (105 lines)
- ❌ `internal/output/*.go` (needs expansion)
- ❌ `cmd/hovimestari/commands/*.go` (8 files) - **LOW PRIORITY**

---

## File Size Analysis

**Large files that may benefit from splitting:**

1. `internal/weather/metno.go` - 442 lines
2. `internal/store/store.go` - 434 lines
3. `internal/config/viper.go` - 384 lines
4. `internal/brief/brief.go` - 350 lines

All four files handle multiple concerns and have refactoring issues filed (#20-23).

---

**Report End**
