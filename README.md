# Dehydrated API Go

A Go package for reading and writing dehydrated's domains.txt files.

## Installation

```bash
go get github.com/schumann-it/dehydrated-api-go
```

## Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/schumann-it/dehydrated-api-go"
)

func main() {
    // Read domains.txt
    entries, err := dehydrated.ReadDomainsFile("domains.txt")
    if err != nil {
        log.Fatal(err)
    }

    // Process entries
    for _, entry := range entries {
        fmt.Printf("Domain: %s, Challenge Types: %v\n", entry.Domain, entry.ChallengeTypes)
    }

    // Create new entries
    newEntries := []dehydrated.DomainEntry{
        {Domain: "example.com"},
        {Domain: "example.org", ChallengeTypes: []string{"dns-01"}},
        {Domain: "example.net", ChallengeTypes: []string{"dns-01", "http-01"}},
    }

    // Write to domains.txt
    if err := dehydrated.WriteDomainsFile("domains.txt", newEntries); err != nil {
        log.Fatal(err)
    }
}
```

## Features

- Read domains.txt files
- Write domains.txt files
- Support for multiple challenge types
- Skip empty lines and comments
- String representation of domain entries

## License

MIT License 