# Redmine Go Library

This Go library provides an interface to interact with the Redmine project management tool.
It supports creating and updating issues, retrieving issue notes, and handling retries for API calls.
Its main focus is on using redmine issues as a ticketing system for customer support.

## Features

- **Dynamic Configuration**: Easily configure the library to connect to your Redmine instance using functional options, and change the configuration at runtime.
- **Issue Management**: Create, update, and check the status of issues in Redmine.
- **Notes Retrieval**: Retrieve notes from issues.
- **Retry Mechanism**: Built-in retry mechanism for API calls to handle transient errors.
- **Logging**: (Optional) Integration with `zerolog` for structured logging.
- **Graceful Shutdown**: Wait for all goroutines to finish before shutting down.

## Installation

To install the library, use `go get`:

```sh
go get gitlab.com/etke.cc/go/redmine
```

## Usage

### Configuration

You need to configure the library to connect to your Redmine instance, here is an example with all available options set:

```go
import (
    "gitlab.com/etke.cc/go/redmine"
    "github.com/rs/zerolog"
    "os"
)

func main() {
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger() // completely optional, by default discard logs
    
    client, err := redmine.New(
        redmine.WithLog(&logger), // optional, by default discard logs
        redmine.WithHost("https://redmine.example.com"), // required
        redmine.WithAPIKey("your_api_key"), // required
        redmine.WithProjectIdentifier("my-project"), // required, the identifier of the project. You may use redmine.WithProjectID() instead
        redmine.WithTrackerID(1), // required
        redmine.WithWaitingForOperatorStatusID(1), // required, the status ID for issues waiting for operator. Usually "In Progress" status
        redmine.WithWaitingForCustomerStatusID(2), // required, the status ID for issues waiting for customer. Usually "New" status
        redmine.WithDoneStatusID(3), // required, the status ID for closed issues. Usually "Closed" status
    )
    
    if err != nil {
        logger.Fatal().Err(err).Msg("Failed to create Redmine client")
    }
    
    // Use the client...
}
```

### Creating an Issue

To create a new issue, use the `NewIssue` method:

```go
issueID, err := client.NewIssue("Issue Subject", "email", "user@example.com", "Issue description text")
if err != nil {
    logger.Error().Err(err).Msg("Failed to create issue")
} else {
    logger.Info().Int64("issue_id", issueID).Msg("Issue created")
}
```

### Updating an Issue

To update the status of an existing issue, use the `UpdateIssue` method:

```go
err := client.UpdateIssue(issueID, redmine.WaitingForCustomer, "Your notes")
if err != nil {
    logger.Error().Err(err).Msg("Failed to update issue")
}
```

### Retrieving Issue Notes

To retrieve the notes of an issue, use the `GetNotes` method:

```go
notes, err := client.GetNotes(issueID)
if err != nil {
    logger.Error().Err(err).Msg("Failed to get issue notes")
} else {
    for _, note := range notes {
        fmt.Println(note.Notes)
    }
}
```

### Checking if an Issue is Closed

To check if an issue is closed, use the `IsClosed` method:

```go
isClosed, err := client.IsClosed(issueID)
if err != nil {
    logger.Error().Err(err).Msg("Failed to check if issue is closed")
} else if isClosed {
    logger.Info().Msg("Issue is closed")
} else {
    logger.Info().Msg("Issue is open")
}
```

### Shutdown

To gracefully shutdown the client and wait for all goroutines to finish, use the `Shutdown` method:

```go
client.Shutdown()
```
