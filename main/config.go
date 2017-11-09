package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type OauthConfig struct {
	Oauth Oauth
}

type Oauth struct {
	Github Github
}

type Github struct {
	Key      string
	Secret   string
	Callback string
}

// LoadConfig returns the configuration of the application
func LoadConfig() OauthConfig {
	file, osErr := os.Open(filepath.Join("config", "config.json"))
	defer file.Close()
	if osErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", osErr)
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	config := OauthConfig{}
	decodeErr := decoder.Decode(&config)
	if decodeErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", decodeErr)
		os.Exit(1)
	}
	return config
}
