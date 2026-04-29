package main

import (
	"fmt"
	"os"

	"github.com/Boopitty/Aggregator/internal/config"
)

func main() {
	config, err := config.Read()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1) // use os.Exit(1) to terminate the program.
		// The rest of the code will not run after this line.
	}

	var s state
	s.cfg = config

	var cmds commands
	cmds.handlers = make(map[string]func(*state, command) error)

	// Register the "login" command with its handler function.
	err = cmds.register("login", handlerLogin)
	if err != nil {
		fmt.Println("Error registering command:", err)
		os.Exit(1)
	}

	// Get the command-line arguments and create a command struct.
	input := os.Args
	if len(input) < 2 {
		fmt.Println("Error: No command provided.")
		os.Exit(1)
	}

	cmdName := input[1:]
	// Create a command struct with the command name and its arguments.
	if len(cmdName) < 2 {
		fmt.Println("Error: No arguments provided for the command.")
		os.Exit(1)
	}
	cmd := command{name: cmdName[0], slice: cmdName[1:]}

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}
}
