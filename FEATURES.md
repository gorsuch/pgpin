## User Features

* "Pins" are persistent records of SQL queries and the results
  of running them against Postgres databases given by users
* Recent pin results are available without re-running the query, and
  to all system users
* Pin results refresh periodically in the background
* Pins can be created against any database for which the user has
  the Postgres URL
* All functionality is available over an HTTP CRUD API

## Implementation Features

* Data stored in Postgres
* Data constraints enforced in Postgres
* Data access via github.com/lib/pq
* Data migration scaffolding
* Data soft deletions
* Data create/update timestamping
* Data optimistic locking
* Data input validation
* Data query results stored in Postgres json type
* Data ids stored in Postgres uuid type
* Data encryption of user database URLs via github.com/fernet/fernet-go
* Data application_name for API and pin queries
* Data statement_timeout and connect_timeout for API and pin queries
* Web API in the style of interagent/http-api-design
* Web request routing via github.com/zenazn/goji/web
* Web request logging
* Web request Ids conveyed in logs and responses
* Web request timeouts
* Web resource dereferencing by id or name
* Web not found handling
* Web error and panic handling
* Web request logging
* Web system status endpoint
* Web endpoints for triggering errors, panics, and timeouts
* Web server graceful shutdown via github.com/zenazn/goji/graceful
* Worker process for user queries outside of HTTP request cycle
* Worker error and panic handling
* Worker user db connection and query error handling
* Worker cool-off prevents spinning on errors or noops
* Worker graceful shutdown
* Config extracted from the Unix environment
* Config validation via github.com/darkhelmet/env
* Logs in key=value style with consistent type keys
* Test exercising full application stack
* Test assertions via github.com/stretchr/testify/assert
* Test workflow documentation
* Godep pegs application dependencies
* Procfile communicates process types
* Development container in Vagrant
* Development workflow instructions
* Development code verification via gofmt, govet, and errcheckto
* Scripting workflow instructions
