package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/Boopitty/Aggregator/internal/config"
	"github.com/Boopitty/Aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	json_file, err := config.Read()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1) // use os.Exit(1) to terminate the program.
		// The rest of the code will not run after this line.
	}

	var s state
	s.cfg = json_file

	// Connect to the database using the URL from the config.
	db, err := sql.Open("postgres", s.cfg.DbUrl)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create a new database.Queries struct using the connected database.
	dbQueries := database.New(db)
	s.db = dbQueries

	var cmds commands
	cmds.handlers = make(map[string]func(*state, command) error)

	// Register the various commands with their handler functions.
	err = cmds.register("login", handlerLogin)
	if err != nil {
		fmt.Println("Error registering command:", err)
		os.Exit(1)
	}

	err = cmds.register("register", handlerRegister)
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
	if strings.TrimSpace(cmd.slice[0]) == "" {
		fmt.Println("Error: Command arguments cannot be empty.")
		os.Exit(1)
	}

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}
}
