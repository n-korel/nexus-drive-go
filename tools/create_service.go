package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	serviceName := flag.String("name", "", "Name of the service (e.g., user, payment)")
	flag.Parse()

	if *serviceName == "" {
		fmt.Println("Please provide a service name using -name flag")
		os.Exit(1)
	}

	// Create service directory structure
	basePath := filepath.Join("services", *serviceName+"-service")
	dirs := []string{
		"cmd",
		"internal/domain",
		"internal/service",
		"internal/infrastructure/events",
		"internal/infrastructure/grpc",
		"internal/infrastructure/repository",
		"pkg/types",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(basePath, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create empty README.md
	readmePath := filepath.Join(basePath, "README.md")
	readmeContent := fmt.Sprintf(`# %s service

This service handles all %s-related operations in the system.
`, *serviceName, *serviceName)

	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		fmt.Printf("Error creating README.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully created %s service structure in %s\n", *serviceName, basePath)
	fmt.Println("\nDirectory structure created:")
	fmt.Printf(`
services/%s-service/
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── domain/           # Business domain models and interfaces
│   │   └── %s.go         # Core domain interfaces
│   ├── service/          # Business logic implementation
│   │   └── service.go    # Service implementations
│   └── infrastructure/   # External dependencies implementations (abstractions)
│       ├── events/       # Event handling (RabbitMQ)
│       ├── grpc/         # gRPC server handlers
│       └── repository/   # Data persistence
├── pkg/                  # Public packages
│   └── types/           # Shared types and models
└── README.md            # This file
`, *serviceName, *serviceName)
} 