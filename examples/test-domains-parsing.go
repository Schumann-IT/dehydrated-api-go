package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/schumann-it/dehydrated-api-go/internal/service"
)

func main() {
	// Path to the example domains.txt file
	domainsFile := filepath.Join("examples", "domains.txt")

	// Read and parse the domains file
	entries, err := service.ReadDomainsFile(domainsFile)
	if err != nil {
		log.Fatalf("Failed to read domains file: %v", err)
	}

	fmt.Printf("Parsed %d domain entries:\n\n", len(entries))

	for i, entry := range entries {
		fmt.Printf("Entry %d:\n", i+1)
		fmt.Printf("  Domain: %s\n", entry.Domain)
		if len(entry.AlternativeNames) > 0 {
			fmt.Printf("  Alternative Names: %v\n", entry.AlternativeNames)
		}
		if entry.Alias != "" {
			fmt.Printf("  Alias: %s\n", entry.Alias)
		}
		fmt.Printf("  Enabled: %t\n", entry.Enabled)
		if entry.Comment != "" {
			fmt.Printf("  Comment: %s\n", entry.Comment)
		}
		fmt.Println()
	}

	// Demonstrate writing back to a file
	outputFile := filepath.Join("examples", "output-domains.txt")
	err = service.WriteDomainsFile(outputFile, entries)
	if err != nil {
		log.Fatalf("Failed to write domains file: %v", err)
	}

	fmt.Printf("Wrote parsed entries back to: %s\n", outputFile)
}
