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
  create-migration \
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
create-migration gm:
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
setup-test:
	go run test-setup/main.go

run-tests t: generate-queries setup-test test-rpcs test-middleware test-service test-permission test-errors test-message test-producer test-helpers test-faults test-logging test-qx


all-tests: setup-test
	# Note: this is not well set up tests have errors. Issue lies on how db connection is used. needs to be fixed
	# most likely need to add a lock to the db connection( somehow )
	# For now use run-tests command
	go test -cover ./... | grep -v 'testutil'

tbreak:
	go test ./... -run "$(t)"

test-rpcs: setup-test
	@go test mist/src/rpcs/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-middleware: setup-test
	@echo -----------------------------------------
	@go test mist/src/middleware/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-service: setup-test
	@echo -----------------------------------------
	@go test mist/src/service/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-permission: setup-test
	@echo -----------------------------------------
	@go test mist/src/permission/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-logging: setup-test
	@echo -----------------------------------------
	@go test mist/src/logging/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)


test-faults: setup-test
	@echo -----------------------------------------
	@go test mist/src/faults/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-message: setup-test
	@echo -----------------------------------------
	@go test mist/src/faults/message/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-producer: setup-test
	@echo -----------------------------------------
	@go test mist/src/producer/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)

test-helpers: setup-test
	@echo -----------------------------------------
	@go test mist/src/helpers/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)


test-qx: setup-test
	@echo -----------------------------------------
	@go test mist/src/psql_db/qx/... -coverprofile=coverage/coverage.out  $(go_test_flags)
	@go tool cover $(go_test_coverage_flags)


# ----- FORMAT -----
lint:
	golangci-lint run -E errcheck

lint-proto:
	@buf lint

# ----- SHORTCUTS -----
psql:
	# Make sure to have all roles for your user
	@psql -U ${DATABASE_ROLE}

update-all-deps:
	go get -u ./...

redis:
	redis-cli -u redis://${REDIS_USERNAME}:${REDIS_PASSWORD}@${REDIS_HOSTNAME}:${REDIS_PORT}
