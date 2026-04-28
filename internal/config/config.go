package config

import (
	"encoding/json"
	"os"
)

const configFileName = "/.gatorconfig.json"

type Config struct {
	// Represents the JSON file structure, including struct tags
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (*Config, error) {
	// reads the JSON file found at ~/.gatorconfig.json and returns a Config struct.
	// It should read the file from the HOME directory,
	// then decode the JSON string into a new Config struct.

	path, err := os.UserHomeDir() // path == "/home/username"
	if err != nil {
		return nil, err
	}
	// Construct the full path to the config file
	configPath := path + configFileName

	// Read the file and decode the JSON into a Config struct
	var config Config
	data, err := os.ReadFile(configPath) // read file and return []byte
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &config) // Decode the JSON data
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) SetUser(name string) error {
	// writes the config struct to the JSON file
	// after setting the current_user_name field.
	c.CurrentUserName = name
	data, err := json.Marshal(c) // Encode the Config struct to JSON
	if err != nil {
		return err
	}

	path, err := os.UserHomeDir() // Get the user's home directory
	if err != nil {
		return err
	}
	configPath := path + configFileName

	err = os.WriteFile(configPath, data, 0644) // Write the JSON data to the file. 0644 is the file permission.
	if err != nil {
		return err
	}
	return nil
}
