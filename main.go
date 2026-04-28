package main

import (
	"fmt"

	"github.com/Boopitty/Aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Println("Config Aquired:", cfg.DbUrl)

	err = cfg.SetUser("Emil")
	if err != nil {
		fmt.Println("Error setting user:", err)
		return
	}
	fmt.Println("User set successfully")

	cfg, err = config.Read()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}
	fmt.Printf("Config Aquired: %s, %s\n", cfg.DbUrl, cfg.CurrentUserName)
}
