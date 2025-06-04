# include variables from the env file
include .envrc


# ============================================================================== #
# HELPERS
# ============================================================================== #


## help: print this help msg
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'


# confirm target
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]



# ============================================================================== #
# DEVELOPMENT
# ============================================================================== #

## run/api: run the cmd/api application/start api
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to the db using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new db migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up db migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up


# ============================================================================== #
# QUALITY CONTROL
# ============================================================================== #

## audit: tidy deps and formst, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...


## vendor: tidy and vendor deps (suits offline needs)
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor


# ============================================================================== #
# BUILD
# ============================================================================== #

current_time = $(shell date --iso-8601=seconds)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'
git_description = $(shell git describe --always --dirty --tags --long)

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api

# ============================================================================== #
# PRODUCTION
# ============================================================================== #
production_host_ip = 'VPS_IP_HERE'

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh greenlight@${production_host_ip}

## production/deploy/api: deploy the api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api greenlight@${production_host_ip}:~
	rsync -rP --delete ./migrations greenlight@${production_host_ip}:~
	rsync -P ./remote/production/api.service greenlight@${production_host_ip}:~
	rsync -P ./remote/production/Caddyfile greenlight@${production_host_ip}:~
	ssh -t greenlight@${production_host_ip} '\
	migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up \
	&& sudo mv ~/api.service /etc/systemd/system/ \
	&& sudo systemctl enable api \
	&& sudo systemctl restart api \
	&& sudo mv ~/Caddyfile /etc/caddy/ \
	&& sudo systemctl reload caddy \
	'
