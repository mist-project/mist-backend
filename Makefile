v := false
html := false
go_test_flags := -tags no_test_coverage
go_test_coverage_flags := -func=coverage/coverage.out

# Add verbose flag if v=true
ifeq ($(v),true)
    go_test_flags += -v
endif

# Add 
ifeq ($(html),true)
    go_test_coverage_flags = -html=coverage/coverage.out
endif

.PHONY: \
  run \
  build \
  live-run \
  compile-protos \
  generate-migration \
  migrate \
  migrate-to \
  migrate-one \
  downgrade \
  downgrade-to \
  dump-schema \
  psql \
  lint

# ----- COMPILE AND EXECUTE -----
run: compile-protos generate-queries
	@go run src/main.go

build: compile-protos generate-queries
	@go build -o bin/mist src/main.go

live-run: compile-protos generate-queries
	@air --build.cmd "go build -o bin/mist src/main.go" --build.bin "./bin/mist"

compile-protos:
	@buf generate

generate-queries:
	@sqlc generate

# ----- DB Migrations -----
generate-migration gm:
	goose create ${message} sql

migrate:
	@goose up
	@make -s dump-schema

migrate-to:
	@goose up-to ${version}
	@make -s dump-schema

migrate-one:
	@goose up-by-one
	@make -s dump-schema

downgrade:
	@goose down || true
	@make -s dump-schema

downgrade-to:
	@goose down-to ${version}  || true
	@make -s dump-schema

dump-schema:
	pg_dump ${DATABASE_NAME} --schema-only | grep -v -e '^--' -e '^COMMENT ON' -e '^REVOKE' -e \
		'^GRANT' -e '^SET' -e 'ALTER DEFAULT PRIVILEGES' -e 'OWNER TO' | cat -s > \
		${DB_SOURCE_DIR}/schema.sql


# ----- TESTS -----
tests t: generate-queries test-service test-middleware

test-service:
	@go test mist/src/rpcs -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-middleware:
	@echo -----------------------------------------
	@go test mist/src/middleware -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

# ----- FORMAT -----
lint:
	golangci-lint run --disable-all -E errcheck

lint-proto:
	@buf lint

# ----- SHORTCUTS -----
psql:
	# Make sure to have all roles for your user
	@psql -U ${DATABASE_ROLE}