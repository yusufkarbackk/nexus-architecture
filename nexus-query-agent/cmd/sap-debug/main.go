package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/SAP/go-hdb/driver"
)

func main() {
	host := flag.String("host", "", "SAP HANA Host")
	port := flag.String("port", "30015", "SAP HANA Port")
	user := flag.String("user", "", "Username")
	password := flag.String("password", "", "Password")
	schema := flag.String("schema", "", "Schema to test (e.g. BATI_TEST)")
	table := flag.String("table", "", "Table to test (e.g. @ATRXMSTR)")

	flag.Parse()

	if *host == "" || *user == "" || *password == "" {
		log.Fatal("Missing required flags: -host, -user, -password")
	}

	// Construct DSN
	dsn := fmt.Sprintf("hdb://%s:%s@%s:%s", *user, *password, *host, *port)
	fmt.Printf("Connecting to %s:%s as %s...\n", *host, *port, *user)

	db, err := sql.Open("hdb", dsn)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Successfully connected to SAP HANA!")

	// 1. List valid schemas for this user
	fmt.Println("\n--- Checking accessible schemas ---")
	rows, err := db.Query("SELECT SCHEMA_NAME FROM SCHEMAS WHERE SCHEMA_NAME = ?", *schema)
	if err != nil {
		fmt.Printf("Error querying schemas: %v\n", err)
	} else {
		defer rows.Close()
		found := false
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err == nil {
				fmt.Printf("Found schema: %s\n", s)
				found = true
			}
		}
		if !found {
			fmt.Printf("WARNING: Schema '%s' not found in SCHEMAS table or not visible to user.\n", *schema)
		}
	}

	// 2. Try simple query
	if *schema != "" && *table != "" {
		// Try fully qualified quoted
		query := fmt.Sprintf("SELECT * FROM \"%s\".\"%s\" LIMIT 1", *schema, *table)
		fmt.Printf("\n--- Executing: %s ---\n", query)

		rows, err := db.Query(query)
		if err != nil {
			log.Fatalf("Query failed: %v", err)
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		fmt.Printf("Columns: %v\n", cols)

		if rows.Next() {
			fmt.Println("Row found!")
		} else {
			fmt.Println("No rows returned (table empty?), but query was successful.")
		}
	}
}
