# confirm target
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


run/api:
	@go run ./cmd/api

# modify if not running locally
db/psql:
	psql ${GREENLIGHT_DB_DSN}

# migrations, with option to pass args just in case of personal testing/stuff
# name=args
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

# include confirm target as prerequisite
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up
