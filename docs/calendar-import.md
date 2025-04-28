# Calendar Import

The Hovimestari application can import calendar events from WebCal URLs (iCalendar format). This document explains how to configure calendar imports and the available update modes.

## Configuration

Calendar sources are configured in the `config.json` file under the `calendars` section:

```json
"calendars": [
  {
    "name": "Family Calendar",
    "url": "webcal://p12-caldav.icloud.com/published/2/...",
    "update_mode": "full_refresh"
  },
  {
    "name": "Work Calendar",
    "url": "webcal://example.com/work-calendar.ics",
    "update_mode": "smart"
  },
  {
    "name": "Holidays in Finland",
    "url": "https://calendars.icloud.com/holidays/fi_fi.ics/"
  }
]
```

## Update Modes

When importing calendar events, Hovimestari offers two different update strategies:

### 1. Full Refresh (`"update_mode": "full_refresh"`)

This is the **default** mode if no `update_mode` is specified.

- **What it does**: Deletes all existing events from this calendar source and reimports them completely.
- **Best for**: Dynamic calendars that change frequently, such as school schedules or family calendars where events might be deleted or moved.
- **Advantages**: Ensures complete accuracy by removing deleted events.
- **Disadvantages**: Less efficient for large calendars that rarely change.

### 2. Smart Update (`"update_mode": "smart"`)

- **What it does**: Updates existing events if they've changed and adds new events, without deleting anything.
- **Best for**: Static calendars that rarely change, such as holidays or namedays.
- **Advantages**: More efficient for large calendars that don't change often.
- **Disadvantages**: Won't remove events that have been deleted from the source calendar.

## Choosing the Right Update Mode

- Use `"full_refresh"` (or omit `update_mode`) for calendars where accuracy is critical and events might be deleted or moved.
- Use `"smart"` for calendars that are mostly static and rarely have events deleted.

## Command Line Usage

To import calendar events manually:

```bash
hovimestari import-calendar
```

This command will import events from all configured calendars using their specified update modes.
