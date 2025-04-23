# SPEC-001: Certificate Management API for Dehydrated

## Background

Dehydrated is a lightweight, bash-based ACME client used to issue and renew Let's Encrypt TLS certificates. While it's
effective in automated environments, it relies on direct access to configuration files like `domains.txt`, making it
less suitable for multi-user or remote infrastructure setups.

Platform engineers often need a centralized and automated way to manage certificates across systems, without manually
editing files or logging into the dehydrated host. To address this, we propose building a REST API layer around
dehydrated, exposing its core functions for external consumption.

This API will power a custom Terraform provider and a single-page web application (SPA), enabling self-service
certificate management without compromising the security or simplicity of dehydrated.

## Requirements

### Must Have

- Provide a REST API to manage entries in the `domains.txt` file (Create, Read, Update, Delete).
- Automatically detect and reflect any manual edits made to the `domains.txt` file.
- Provide a mechanism to reload the domains list from disk upon change (via polling or file watcher).
- Return up-to-date domain information on every API request.
- Be stateless with respect to certificate generation; only manages `domains.txt`.

### Should Have

- Terraform provider that uses the REST API to create/delete domain entries.
- SPA dashboard for managing domain entries in `domains.txt`.

### Could Have

- Trigger dehydrated runs via API (e.g. issue certs).
- Support comments or tags for entries (to identify who owns a cert).

### Won’t Have (yet)

- Authentication or authorization (public API for internal trusted use only).
- Certificate revocation or renewal logic in the API (delegated to dehydrated/cron).

## Method

The system is designed with three main components:

1. A REST API service running **on the same host as dehydrated**, directly manipulating the `domains.txt` file.
2. A custom **Terraform provider** that communicates with the REST API.
3. A **Single Page App (SPA)** dashboard that interacts with the API via HTTP.

The API watches or reloads `domains.txt` on changes and exposes a CRUD interface. It does not handle certificate
issuance directly, leaving that to scheduled dehydrated runs (e.g., via cron).

### Architecture Diagram

```plantuml
@startuml
actor "User (Platform Engineer)" as User

package "Cert Management Host" {
  component "Dehydrated" as D
  component "Domain REST API" as API
  file "domains.txt" as TXT

  API --> TXT : Read/Write
  D --> TXT : Read
}

package "Remote Tools" {
  component "Terraform Provider" as TF
  component "SPA Dashboard" as SPA

  TF --> API : HTTP CRUD
  SPA --> API : HTTP CRUD
}

User --> SPA : Manage domains
User --> TF : Plan/apply infrastructure
@enduml
