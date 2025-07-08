package main

import (
	"fmt"
	"os"

	"github.com/oklog/ulid/v2"
)

func createMigration(name string) error {
	id := ulid.Make().String()
	base := fmt.Sprintf("migrations/%s_%s", id, name)
	upFile := base + ".up.sql"
	downFile := base + ".down.sql"

	if err := os.WriteFile(upFile, []byte("-- Write your UP migration here\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(downFile, []byte("-- Write your DOWN migration here\n"), 0644); err != nil {
		return err
	}
	fmt.Printf("Created migration files:\n  %s\n  %s\n", upFile, downFile)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrator <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  create <name>   Create a new migration")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: migrator create <name>")
			os.Exit(1)
		}
		name := os.Args[2]
		if err := createMigration(name); err != nil {
			fmt.Printf("Error creating migration: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
