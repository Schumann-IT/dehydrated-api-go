package service

import (
	"bufio"
	"os"
	"strings"

	pb "github.com/schumann-it/dehydrated-api-go/plugin/proto"

	"github.com/schumann-it/dehydrated-api-go/internal/model"
)

// ReadDomainsFile reads a domains.txt file and returns a slice of DomainEntry.
// It parses the file format which supports:
// - Domain names with optional alternative names
// - Aliases using the '>' syntax
// - Comments using '#' prefix or inline
// - Disabled entries (prefixed with '#')
func ReadDomainsFile(filename string) ([]model.DomainEntry, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.DomainEntry{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var entries []model.DomainEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		// Check if the line is a comment
		enabled := true
		comment := ""
		if strings.HasPrefix(line, "#") {
			// Remove the comment marker
			line = strings.TrimPrefix(line, "#")
			line = strings.TrimSpace(line)
			enabled = false
		}

		// Extract inline comment if present
		if strings.Contains(line, "#") {
			parts := strings.SplitN(line, "#", 2)
			line = strings.TrimSpace(parts[0])
			comment = strings.TrimSpace(parts[1])
		}

		// Split by '>' to handle aliases
		parts := strings.Split(line, ">")
		mainPart := strings.TrimSpace(parts[0])
		alias := ""
		if len(parts) > 1 {
			alias = strings.TrimSpace(parts[1])
		}

		// Split the main part into domain and alternative names
		fields := strings.Fields(mainPart)
		if len(fields) == 0 {
			continue
		}

		entry := model.DomainEntry{
			DomainEntry: pb.DomainEntry{
				Domain:           fields[0],
				AlternativeNames: fields[1:],
				Alias:            alias,
				Enabled:          enabled,
				Comment:          comment,
			},
		}

		// Only add valid domain entries
		if model.IsValidDomainEntry(&entry) {
			entries = append(entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}

// WriteDomainsFile writes a slice of DomainEntry to a domains.txt file.
// It formats each entry according to the dehydrated domains.txt format:
// - Disabled entries are prefixed with '#'
// - Alternative names are space-separated
// - Aliases are added with ' > ' separator
// - Comments are added with ' # ' separator
func WriteDomainsFile(filename string, entries []model.DomainEntry) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, entry := range entries {
		// Build the line
		var line strings.Builder

		// Add comment marker if disabled
		if !entry.Enabled {
			line.WriteString("# ")
		}

		// Add domain and alternative names
		line.WriteString(entry.Domain)
		for _, altName := range entry.AlternativeNames {
			line.WriteString(" ")
			line.WriteString(altName)
		}

		// Add alias if present
		if entry.Alias != "" {
			line.WriteString(" > ")
			line.WriteString(entry.Alias)
		}

		// Add comment if present
		if entry.Comment != "" {
			line.WriteString(" # ")
			line.WriteString(entry.Comment)
		}

		// Write the line
		if _, err := writer.WriteString(line.String() + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}
