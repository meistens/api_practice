run:
	@go run ./cmd/api

# modify if not running locally
psql:
	psql ${GREENLIGHT_DB_DSN}

up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up
