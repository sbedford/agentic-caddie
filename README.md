# agentic-caddie

## Commands for build/test

### Build Commands
* go build -o bin/api cmd/api/main.go
* go build -o bin/bootstrap cmd/bootstrap/main.go
* go build -o bin/agent cmd/agent/main.go

### Run Tests
 go test internal/models/*

### Generate SQLC
sqlc generate - check unicode characters

### Update Swagger Definition
swag init -g cmd/api/main.go -o cmd/api/docs

### Run Web Server
npm run dev (in web folder)