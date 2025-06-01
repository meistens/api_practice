run:
	@go run ./cmd/api

# modify if not running locally
psql:
	psql ${GREENLIGHT_DB_DSN}

# migrations, with option to pass args just in case of personal testing/stuff
# name=args
migration:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up
