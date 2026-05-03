package main

import (
	"database/sql"
	"fmt"
	"os"

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

	err = cmds.register("reset", handlerReset)
	if err != nil {
		fmt.Println("Error registering command:", err)
		os.Exit(1)
	}

	err = cmds.register("users", handlerGetAll)
	if err != nil {
		fmt.Println("Error registering command:", err)
		os.Exit(1)
	}

	err = cmds.register("agg", agg)
	if err != nil {
		fmt.Println("Error registering command:", err)
		os.Exit(1)
	}

	err = cmds.register("addfeed", handlerAddFeed)
	if err != nil {
		fmt.Println("Error registering command:", err)
		os.Exit(1)
	}

	err = cmds.register("feeds", handlerFeeds)
	if err != nil {
		fmt.Println("Error registering command:", err)
		os.Exit(1)
	}

	// Get the command-line arguments and create a command struct.
	fullInput := os.Args
	if len(fullInput) < 2 {
		fmt.Println("Error: No command provided.")
		os.Exit(1)
	}

	input := fullInput[1:]
	cmd := command{name: input[0], slice: nil}
	if len(input) > 1 {
		cmd.slice = input[1:]
	}

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}
}
