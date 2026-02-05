package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"nexus-query-agent/internal/config"
	"nexus-query-agent/internal/connection"
	"nexus-query-agent/internal/executor"
	"nexus-query-agent/internal/models"
)

func main() {
	// Parse flags
	configPath := flag.String("config", "config/config.yml", "Path to config file")
	flag.Parse()

	log.Println("===========================================")
	log.Println("       Nexus Query Agent Starting          ")
	log.Println("===========================================")

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("ERROR: Failed to load config: %v", err)
	}

	log.Printf("INFO: Agent ID: %s", cfg.Agent.ID)
	log.Printf("INFO: Agent Name: %s", cfg.Agent.Name)
	log.Printf("INFO: Nexus Core URL: %s", cfg.Nexus.CoreURL)

	// Create Nexus client
	client := connection.NewNexusClient(cfg)

	// Set query handler - connections are now dynamic per-request
	client.OnQueryRequest = func(req *models.QueryRequest) {
		handleQueryRequest(client, cfg, req)
	}

	// Connect to Nexus Core
	if err := client.Connect(); err != nil {
		log.Printf("ERROR: Failed to connect to Nexus Core: %v", err)
		log.Println("INFO: Will retry connection...")
		go client.Reconnect()
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("INFO: Query Agent is running. Press Ctrl+C to stop.")

	<-sigChan

	log.Println("INFO: Shutting down...")
	client.Close()
	log.Println("INFO: Shutdown complete")
}

// handleQueryRequest processes incoming query requests with dynamic connections
func handleQueryRequest(client *connection.NexusClient, cfg *config.Config, req *models.QueryRequest) {
	log.Printf("INFO: Processing %s request %s for datasource %s:%d",
		req.QueryType, req.RequestID, req.Datasource.Host, req.Datasource.Port)

	// Create executor based on datasource type
	exec, err := executor.NewExecutor(req.Datasource.Type, &cfg.Limits)
	if err != nil {
		client.SendError(req.RequestID, "UNSUPPORTED_DATASOURCE", err.Error())
		return
	}

	var result *models.QueryResult

	// Route based on query type
	queryType := req.QueryType
	if queryType == "" {
		queryType = "select" // Default to SELECT for backward compatibility
	}

	switch queryType {
	case "select":
		// Execute SELECT query with pagination
		result, err = exec.Execute(&req.Datasource, req.Query, req.Page, req.Limit)
	case "insert", "update", "delete":
		// Execute DML with transaction handling
		if sapExec, ok := exec.(*executor.SapExecutor); ok {
			result, err = sapExec.ExecuteDML(&req.Datasource, queryType, req.Query, req.Params)
		} else {
			client.SendError(req.RequestID, "DML_NOT_SUPPORTED", "DML operations only supported for SAP datasources")
			return
		}
	default:
		client.SendError(req.RequestID, "INVALID_QUERY_TYPE", "Query type must be: select, insert, update, or delete")
		return
	}

	if err != nil {
		client.SendError(req.RequestID, "EXECUTION_ERROR", err.Error())
		return
	}

	result.RequestID = req.RequestID

	// Send result
	if err := client.SendResult(result); err != nil {
		log.Printf("ERROR: Failed to send result: %v", err)
		return
	}

	// Log appropriate message based on query type
	if queryType == "select" {
		log.Printf("INFO: Query %s completed in %dms, %d rows returned",
			req.RequestID, result.ExecutionTimeMs, len(result.Data))
	} else {
		log.Printf("INFO: %s %s completed in %dms, %d rows affected",
			queryType, req.RequestID, result.ExecutionTimeMs, result.AffectedRows)
	}
}
