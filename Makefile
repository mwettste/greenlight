include .envrc

# ============================================================================= #
# HELPERS
# ============================================================================= #
## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
	
confirm:
	@echo -n 'Are you sure [y/N] ' && read ans && [ $${ans:-N} = y ]

# ============================================================================= #
# DEVELOPMENT 
# ============================================================================= #
## run/api: run the cmd/api application
run/api:
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN} -smtp-username=${MAILTRAP_USER} -smtp-password=${MAILTRAP_PW}

## db/psql: connect to the database using psql
db/psql:
	docker exec -it deploy_db_1 bash -c "su - postgres"

## db/migrations/new name=$1: create a new database migration
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

# ============================================================================= #
# QUALITY CONTROL
# ============================================================================= #
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# ============================================================================= #
# BUILD
# ============================================================================= #
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api