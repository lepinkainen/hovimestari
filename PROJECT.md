# Hovimestari — Project Purpose

## What it is

A personal AI butler ("hovimestari" is Finnish for butler), directly inspired by Geoffrey Litt's ["How I Made a Useful AI Assistant with One SQLite Table and a Handful of Cron Jobs"](https://www.geoffreylitt.com/2025/04/12/how-i-made-a-useful-ai-assistant-with-one-sqlite-table-and-a-handful-of-cron-jobs).

The architecture is the whole idea: **one SQLite table of "memories"**, a handful of importers that write into it, and a generator that turns the relevant subset into a daily brief. No agent framework, no vector database, no orchestration layer.

## Current capabilities

- Calendar import from multiple iCloud/WebCal feeds, with smart and full-refresh modes
- Weather import from the MET Norway API
- Family information — members, birthdays
- Lunch menu integration via the `palmia-lunch` package
- Daily briefs generated with Google Gemini, written in Finnish in a formal butler register
- Output to CLI, Discord webhooks and Telegram bots
- Pure Go, CGO-free, cross-compiles cleanly to Linux AMD64 for the server that runs the cron jobs

## Direction

Add importers, not machinery. Any new capability should be a source that writes memories plus a prompt change — if it needs new infrastructure, it probably does not belong.

## Constraints

- **Two SQLite tables: `memories` and `calendar_events`.** Memories stay one table; structured calendar data is the one exception. Resist further schema sprawl; that simplicity is the design.
- Pure Go / no cgo, so `task build-linux` produces a static binary.
- Tests must stay deterministic with no external dependencies.
- Briefs are in Finnish with a consistent butler tone — prompts live in `prompts.json`.
- Historical note: the project was largely LLM-written (Cline) up to v1.0; the README documents that the author moved to hand-coding from there.
