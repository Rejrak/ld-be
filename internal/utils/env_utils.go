package utils

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path/filepath"
)

// EnvUtil is a utility struct for handling environment variable loading.
type EnvUtil struct{}

// LoadEnv loads environment variables from the specified .env file.
// It panics if the file cannot be loaded, providing a clear error message.
func (e EnvUtil) LoadEnv(envFile string) {
	// Load environment variables from the specified .env file
	err := godotenv.Load(e.dir(envFile))
	if err != nil {
		panic(fmt.Errorf("Error loading .env file: %w", err)) // Panic if there is an error loading the file
	}
}

// dir determines the directory path to the .env file by traversing up the directory structure.
// It looks for the presence of a "go.mod" file as a marker of the project root.
func (e EnvUtil) dir(envFile string) string {
	currentDir, err := os.Getwd() // Get the current working directory
	if err != nil {
		panic(err) // Panic if there is an error retrieving the current directory
	}

	// Traverse up the directory structure to locate the project root (identified by "go.mod")
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break // Stop when "go.mod" is found, indicating the project root
		}

		// Move to the parent directory if "go.mod" is not found
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found")) // Panic if the project root cannot be determined
		}
		currentDir = parent
	}

	// Return the full path to the specified .env file in the project root directory
	return filepath.Join(currentDir, envFile)
}

// Env is a global instance of EnvUtil, used to load environment variables throughout the application.
var (
	Env = EnvUtil{}
)
