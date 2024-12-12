# nxs-go-redmine

## Introduction

Go client library for [Redmine](https://www.redmine.org)

### Features

- Compatible with Redmine 4.2+ (*some functionality provided by this library may not be available when working with Redmine versions below 5.0*)
- Implemented following Redmine resources:
  - [Issues](https://www.redmine.org/projects/redmine/wiki/Rest_Issues)
  - [Projects](https://www.redmine.org/projects/redmine/wiki/Rest_Projects)
  - [Project Memberships](https://www.redmine.org/projects/redmine/wiki/Rest_Memberships)
  - [Users](https://www.redmine.org/projects/redmine/wiki/Rest_Users)
  - [Time Entries](https://www.redmine.org/projects/redmine/wiki/Rest_TimeEntries)
  - [Wiki Pages](https://www.redmine.org/projects/redmine/wiki/Rest_WikiPages)
  - [Attachments](https://www.redmine.org/projects/redmine/wiki/Rest_Attachments)
  - [Issue Statuses](https://www.redmine.org/projects/redmine/wiki/Rest_IssueStatuses)
  - [Trackers](https://www.redmine.org/projects/redmine/wiki/Rest_Trackers)
  - [Enumerations](https://www.redmine.org/projects/redmine/wiki/Rest_Enumerations)
  - [Groups](https://www.redmine.org/projects/redmine/wiki/Rest_Groups)
  - [Custom Fields](https://www.redmine.org/projects/redmine/wiki/Rest_CustomFields)

### New in nxs-go-redmine v5

- All implemented method updated to work with latest Redmine API version
- All fields that may be omitted are now pointers
- Replaced all IDs from `int` to `int64`
- Includes for methods now are specified with named constants
- Added tools to operate with filters and sorts

### Who can use the tool

Developer teams or sysadmins who need to automate a business processes that works around Redmine.

## Quickstart

### Import

```go
import "github.com/nixys/nxs-go-redmine/v5"
```

### Initialize

To initialize this library you need to do:
- Initialize context via call `redmine.Init()` function with specified values of Redmine endpoint and API key

After thar you be able to use all available methods to interact with Redmine API.

## Example

In the example below will be printed a names for all active projects from Redmine

```go
package main

import (
	"fmt"
	"os"

	redmine "github.com/nixys/nxs-go-redmine/v5"
)

func main() {

	// Get variables from environment for connect to Redmine server
	rdmnHost := os.Getenv("REDMINE_HOST")
	rdmnAPIKey := os.Getenv("REDMINE_API_KEY")
	if rdmnHost == "" || rdmnAPIKey == "" {
		fmt.Println("Init error: make sure environment variables `REDMINE_HOST` and `REDMINE_API_KEY` are defined")
		os.Exit(1)
	}

	r := redmine.Init(
		redmine.Settings{
			Endpoint: rdmnHost,
			APIKey:   rdmnAPIKey,
		},
	)

	fmt.Println("Init: success")

	// Get all active projects with additional 
	// fields (trackers, categories and modules)
	p, _, err := r.ProjectAllGet(
		redmine.ProjectAllGetRequest{
			Includes: []redmine.ProjectInclude{
				redmine.ProjectIncludeTrackers,
				redmine.ProjectIncludeIssueCategories,
				redmine.ProjectIncludeEnabledModules,
			},
			Filters: redmine.ProjectGetRequestFiltersInit().
				StatusSet(redmine.ProjectStatusActive),
		})
	if err != nil {
		fmt.Println("Projects get error:", err)
		os.Exit(1)
	}

	fmt.Println("Projects:")
	for _, e := range p.Projects {
		fmt.Println("-", e.Name)
	}
}

```

Run:

```
REDMINE_HOST="https://redmine.yourdomain.com" REDMINE_API_KEY="YOUR_API_KEY" go run main.go
```

**For more examples see apps based on this library:**
- [nxs-support-bot](https://github.com/nixys/nxs-support-bot)

## Roadmap

Following features are already in backlog for our development team and will be released soon:
- Implement more Redmine API methods (let us know which one you want to see at first)
- Improve error handling in the library

## Feedback

For support and feedback please contact me:
- [Issues](https://github.com/nixys/nxs-go-redmine/issues)
- telegram: [@borisershov](https://t.me/borisershov)
- e-mail: b.ershov@nixys.ru

## License

nxs-go-redmine is released under the [MIT License](LICENSE).
