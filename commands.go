package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Boopitty/Aggregator/internal/config"
	"github.com/Boopitty/Aggregator/internal/database"
	"github.com/google/uuid"
)

// The state is the user's current states,
// such as the current user, the current session, etc.
type state struct {
	cfg *config.Config
	db  *database.Queries
}

// The command struct represents a command that the user can run,
type command struct {
	name  string
	slice []string
}

// The commands struct holds a map of command names to their handler functions.
type commands struct {
	handlers map[string]func(*state, command) error
}

// This will be the function signature of all command handlers.
func handlerLogin(s *state, cmd command) error {
	if cmd.slice == nil {
		return fmt.Errorf("No name provided")
	}
	name := cmd.slice[0]

	_, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		fmt.Println("User not found")
		os.Exit(1)
	}
	// Set the user to the fist element of the slice, which should be the username.
	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("Error setting user: %w", err)
	}

	fmt.Println("User set successfully")
	return nil
}

// Register a new user into the database and set the current user to the new user's name.
func handlerRegister(s *state, cmd command) error {
	// This is a placeholder for the register command handler.
	if cmd.slice == nil {
		return fmt.Errorf("No name provided")
	}

	name := cmd.slice[0]
	// Check if the user already exists in the database.
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {
		fmt.Println("User already exists")
		os.Exit(1)
	}

	// Create a new user in the database with the provided name.
	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("Error creating user: %w", err)
	}

	// Set the current user in the config to the new user's name.
	s.cfg.SetUser(newUser.Name)
	fmt.Printf(
		"User registered successfully:\nID: %s\nName: %s\nCreated At: %s\nUpdated At: %s\n",
		newUser.ID, newUser.Name, newUser.CreatedAt, newUser.UpdatedAt,
	)
	return nil
}

// Empty the contents of the database
func handlerReset(s *state, cmd command) error {
	// This is a placeholder for the reset command handler.
	err := s.db.Reset(context.Background())
	if err != nil {
		fmt.Println("Error resetting database: %w", err)
		os.Exit(1)
	}
	fmt.Println("Database reset successfully")
	return nil
}

// Get the names of all users and mark the current user
func handlerGetAll(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("Error getting users:", err)
		os.Exit(1)
	}
	for _, user := range users {
		if user == s.cfg.CurrentUserName {
			fmt.Printf("%s (current)\n", user)
		} else {
			fmt.Println(user)
		}
	}
	return nil
}

func agg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("Error fetching feed: %w", err)
	}
	fmt.Println(feed)
	return nil
}

// Runs a given command with the provided state if it exists.
func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("Command not found: %s", cmd.name)
	}
	return handler(s, cmd) // Run the func and return the error value it produces.
}

// Registers a new handler function for a command name.
func (c *commands) register(name string, f func(*state, command) error) error {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = f
	return nil
}
