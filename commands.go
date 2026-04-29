package main

import (
	"fmt"

	"github.com/Boopitty/Aggregator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name  string
	slice []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

// This will be the function signature of all command handlers.
func handlerLogin(s *state, cmd command) error {
	if cmd.slice == nil {
		return fmt.Errorf("No slice provided")
	}

	// Set the user to the fist element of the slice, which should be the username.
	err := s.cfg.SetUser(cmd.slice[0])
	if err != nil {
		return fmt.Errorf("Error setting user: %w", err)
	}

	fmt.Println("User set successfully")
	return nil
}

// Runs a given command with the provided state if it exists.
func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("Command not found: %s", cmd.name)
	}
	// Call the handler function with the state and command.
	// The handler function runs and the result, the error value, is returned.
	return handler(s, cmd)
}

// Registers a new handler function for a command name.
func (c *commands) register(name string, f func(*state, command) error) error {
	// Handle a nil map if needed, then add the handler
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = f
	return nil
}
