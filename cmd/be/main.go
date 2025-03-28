package main

import (
	servConfig "be/internal/config"
	"be/internal/database/db"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"goa.design/clue/log"
)

// main is the entry point for the application.
// It loads the server configuration, initializes services, and handles environment-specific setups.
func main() {
	srvConf := servConfig.LoadServerConfig() // Load server configuration settings from a config file or environment variables.

	// Set up logging format based on whether the output is a terminal or a file.
	format := log.FormatJSON
	if log.IsTerminal() {
		format = log.FormatText
	}

	// Create a context for the application with logging format and debug settings.
	ctx := log.Context(context.Background(), log.WithFormat(format))
	if srvConf.Debug {
		ctx = log.Context(ctx, log.WithDebug()) // Enable debug logs if debug mode is set
		log.Debugf(ctx, "debug logs enabled")
	}

	var wg sync.WaitGroup    // WaitGroup to manage goroutines
	errc := make(chan error) // Error channel to listen for fatal errors

	go handleSignals(errc) // Start goroutine to listen for OS signals (e.g., SIGINT, SIGTERM)

	if err := moveFile("./gen/http/openapi3.yaml", "./static/openapi3.yaml"); err != nil {
		log.Debugf(ctx, "error: %v", err)
	}

	// Create a cancellable context to manage server shutdown.
	ctx, cancel := context.WithCancel(ctx)

	// Set up environment-specific configurations
	switch srvConf.Domain {
	case "development":
		db.ConnectDb()
		epsMap := servConfig.InitializeServices(ctx)               // Initialize and map services to endpoints
		u := srvConf.BuildServerURL(srvConf, ctx)                  // Build server URL based on configuration
		HandleHttpServer(ctx, u, &wg, errc, srvConf.Debug, epsMap) // Start the HTTP server for development

	case "production":
		db.ConnectDb()                                             // Connect to the database for production
		epsMap := servConfig.InitializeServices(ctx)               // Initialize and map services to endpoints
		u := srvConf.BuildServerURL(srvConf, ctx)                  // Build server URL based on configuration
		HandleHttpServer(ctx, u, &wg, errc, srvConf.Debug, epsMap) // Start the HTTP server for production

	default:
		log.Fatal(ctx, fmt.Errorf("invalid host argument: %q (valid hosts: development|production)", srvConf.Domain)) // Fatal error for invalid domain
	}

	// Wait for an error or signal to exit.
	log.Printf(ctx, "exiting (%v)", <-errc)
	cancel()                  // Cancel context to begin shutdown process
	wg.Wait()                 // Wait for all goroutines to complete
	log.Printf(ctx, "exited") // Log when the application has fully exited
}

// handleSignals listens for OS signals and sends them to the error channel.
// This function enables graceful shutdown on system signals (e.g., SIGINT, SIGTERM).
func handleSignals(errc chan error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM) // Notify on interrupt or terminate signals
	errc <- fmt.Errorf("%s", <-c)                     // Send the received signal to the error channel as a formatted error
}

// moveFile moves a file from src to dst. If the destination file exists, it will be overwritten.
func moveFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w or already moved", err)
	}

	err = os.WriteFile(dst, input, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}
