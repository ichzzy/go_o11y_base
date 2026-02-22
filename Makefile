PROJECT_NAME := go_o11y_base
PROJECT_NAME_HYPHEN := go-o11y-base

.PHONY: install-bin uninstall-bin
BIN_DIR := $(CURDIR)/bin
$(BIN_DIR): ; mkdir -p $(BIN_DIR)
$(BIN_DIR)/%: | $(BIN_DIR) ; GOBIN=$(BIN_DIR) go install $(PACKAGE)
$(BIN_DIR)/mockgen:       PACKAGE=go.uber.org/mock/mockgen@v0.6.0
$(BIN_DIR)/golangci-lint: PACKAGE=github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6
$(BIN_DIR)/migrate:       PACKAGE=-tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.2

install-bin: | $(BIN_DIR)/mockgen $(BIN_DIR)/golangci-lint $(BIN_DIR)/migrate
uninstall-bin: ; rm -rf $(BIN_DIR)

.PHONY: test test-race
test:
	go test -v ./internal/... -cover -coverpkg ./internal/... -coverprofile=coverage.out
test-race:
	go test -v ./internal/... -race -cover -coverpkg ./internal/...

.PHONY: lint
lint:
	@$(BIN_DIR)/golangci-lint run ./cmd/... ./internal/... --config=./.golangci.yaml

# Export: curl --location 'http://localhost:6060/debug/pprof/heap' --output heap.out
# Use:    go tool pprof -http=:8080 heap.out
.PHONY: pprof-heap pprof-profile pprof-goleak
pprof-heap:
	@go tool pprof -http=:8077 http://localhost:6060/debug/pprof/heap
pprof-profile:
	@go tool pprof -http=:8078 "http://localhost:6060/debug/pprof/profile?seconds=30"
pprof-goleak:
	@go tool pprof -http=:8079 "http://localhost:6060/debug/pprof/goroutine"

.PHONY: mysql-up mysql-down mysql-new
MYSQL_SQL_PATH = "./migration/mysql"
mysql-init:
	@mysql -h 127.0.0.1 -u root -P 3306 -e "CREATE DATABASE IF NOT EXISTS \`$(PROJECT_NAME)\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;"
mysql-up:
	@$(BIN_DIR)/migrate --database "mysql://root:@tcp(localhost:3306)/$(PROJECT_NAME)?charset=utf8mb4&parseTime=True&loc=Local" --path ${MYSQL_SQL_PATH} up
mysql-down:
	@$(BIN_DIR)/migrate --database "mysql://root:@tcp(localhost:3306)/$(PROJECT_NAME)?charset=utf8mb4&parseTime=True&loc=Local" --path ${MYSQL_SQL_PATH} down -all
mysql-new:
	@( \
		printf "Enter migrate file name: "; read -r MIGRATE_NAME && \
		$(BIN_DIR)/migrate create -ext sql -dir ${MYSQL_SQL_PATH} $${MIGRATE_NAME} \
	)

.PHONY: build-image
build-image:
	docker build \
		-f ./build/Dockerfile \
		-t $(PROJECT_NAME_HYPHEN) \
		--platform linux/amd64 \
		.

.PHONY: clean
clean:
	@rm -f logs/log*
	@rm -f cmd/server/__debug_bin*
	@rm -f *.out

.PHONY: mock
MOCK_SOURCE_DIR := $(CURDIR)/internal/gen/mock
mock:
	rm -rf $(MOCK_SOURCE_DIR)/*_mock.go
	$(BIN_DIR)/mockgen -package mock -source $(CURDIR)/internal/domain/usecase/auth.go -destination $(MOCK_SOURCE_DIR)/auth_uc_mock.go
	$(BIN_DIR)/mockgen -package mock -source $(CURDIR)/internal/domain/usecase/rbac.go -destination $(MOCK_SOURCE_DIR)/rbac_uc_mock.go
	$(BIN_DIR)/mockgen -package mock -source $(CURDIR)/internal/domain/repository/rbac.go -destination $(MOCK_SOURCE_DIR)/rbac_repo_mock.go
	$(BIN_DIR)/mockgen -package mock -source $(CURDIR)/internal/domain/repository/redis_client.go -destination $(MOCK_SOURCE_DIR)/redis_client_mock.go
	$(BIN_DIR)/mockgen -package mock -source $(CURDIR)/internal/domain/repository/user.go -destination $(MOCK_SOURCE_DIR)/user_repo_mock.go