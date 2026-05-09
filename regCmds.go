package main

import (
	"fmt"
)

// Register the various commands with their handler functions into the commands struct.
// This allows for easy management and additions of commands.
func registerCommandHandlers(cmds *commands) error {

	err := cmds.register("login", handlerLogin)
	if err != nil {
		return fmt.Errorf("Error registering login command: %w", err)
	}

	err = cmds.register("register", handlerRegister)
	if err != nil {
		return fmt.Errorf("Error registering register command: %w", err)
	}

	err = cmds.register("reset", handlerReset)
	if err != nil {
		return fmt.Errorf("Error registering reset command: %w", err)
	}

	err = cmds.register("users", handlerGetAll)
	if err != nil {
		return fmt.Errorf("Error registering users command: %w", err)
	}

	err = cmds.register("agg", agg)
	if err != nil {
		return fmt.Errorf("Error registering agg command: %w", err)
	}

	err = cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	if err != nil {
		return fmt.Errorf("Error registering addfeed command: %w", err)
	}

	err = cmds.register("feeds", handlerFeeds)
	if err != nil {
		return fmt.Errorf("Error registering feeds command: %w", err)
	}

	err = cmds.register("follow", middlewareLoggedIn(handlerFollow))
	if err != nil {
		return fmt.Errorf("Error registering follow command: %w", err)
	}

	err = cmds.register("following", middlewareLoggedIn(handlerFollowing))
	if err != nil {
		return fmt.Errorf("Error registering following command: %w", err)
	}

	err = cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	if err != nil {
		return fmt.Errorf("Error registering unfollow command: %w", err)
	}

	err = cmds.register("browse", middlewareLoggedIn(handlerBrowse))
	if err != nil {
		return fmt.Errorf("Error registering browse command: %w", err)
	}
	return nil
}
