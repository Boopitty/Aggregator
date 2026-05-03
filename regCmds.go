package main

import (
	"fmt"
	"os"
)

// Register the various commands with their handler functions into the commands struct.
// This allows for easy management and additions of commands.
func registerCommandHandlers(cmds *commands) error {

	err := cmds.register("login", handlerLogin)
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
	return nil
}
