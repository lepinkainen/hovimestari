package main

// Context represents the shared context for all Kong commands
type Context struct {
	// Shared context for all commands
	// This can be extended with commonly used objects like store, config, etc.
	// For now, it's empty as commands handle their own dependencies
}