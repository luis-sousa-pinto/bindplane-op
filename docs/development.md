# Development

## Requirements
Ensure that the following tools are installed:
- Go 1.20 ([install instructions](https://go.dev/doc/install))
- Node 16.x.x ([install instructions](https://nodejs.dev/learn/how-to-install-nodejs))
- GoReleaser ([install instructions](https://goreleaser.com/install/))

Run the following commands after a fresh clone:
- `make install-tools` (installs development tools)
- `make init-server` (initializes the server config)

## Local Build
To run a local build:
1. Run `make dev`
2. Navigate to `http://localhost:3000`

## Testing
The following commands are used to run tests:
1. `make test` (runs server tests)
2. `make ui-test` (runs ui tests)

## Other Commands
- `make build` (creates a build of bindplane-op)
- `make swagger` (creates REST API swagger documentation)
- `make generate` (generates the graphql server)
- `make help` (lists make targets with descriptions)
