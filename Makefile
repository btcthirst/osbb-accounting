# =========================================
#	OSBB ACCOUNTING MAKEFILE
# =========================================
APP_NAME = osbb-accounting
MAIN = ./cmd/main.go
BUILD_DIR = ./build
BIN = $(BUILD_DIR)/$(APP_NAME)
DB_FILE = osbb.db

# =========================================
#	MAIN COMMANDS
# =========================================

.PHONY: run
run: ## Запускає додаток напряму з вихідних файлів
	@echo "🚀 Запуск $(APP_NAME)..."
	go run $(MAIN)

.PHONY: build
build: ## Компілює додаток у build/osbb-accounting
	@echo "🔧 Збірка бінарника..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BIN) $(MAIN)
	@echo "✅ Бінарник створено: $(BIN)"

.PHONY: clean
clean: ## Видаляє збірки та тимчасові файли
	@echo "🧹 Очищення..."
	rm -rf $(BUILD_DIR) $(DB_FILE) coverage.out
	@echo "✅ Готово!"

.PHONY: example
example: ## Додавання тестових даних
	go run ./cmd/example/main.go

# =========================================
#	МІГРАЦІЇ
# =========================================

.PHONY: migrate
migrate:
	go run ./cmd/migrate/main.go migrate

.PHONY: migrate-rollback
migrate-rollback:
	go run ./cmd/migrate/main.go rollback

.PHONY: migrate-validate
migrate-validate:
	go run ./cmd/migrate/main.go validate

.PHONY: db-version
db-version:
	go run ./cmd/migrate/main.go version

.PHONY: migrate-repair
migrate-repair: ## Швидке виправлення проблеми
	go run ./cmd/migrate/main.go repair

.PHONY: migrate-reset
migrate-reset: ## Повнe скидання БД (ВИДАЛЯЄ ВСІ ДАНІ!)
	go run ./cmd/migrate/main.go reset

.PHONY: migrate-status
migrate-status: ## Статус міграцій
	go run ./cmd/migrate/main.go status

# ===============================
#   DEV MODE (Manual Rebuild & Run)
# ===============================

.PHONY: dev
dev: ## Пересобирає та запускає проєкт вручну
	@echo "🧩 Режим розробника..."
	@mkdir -p $(BUILD_DIR)
	@echo "🔧 Збірка..."
	go build -o $(BIN) $(MAIN)
	@echo "🚀 Запуск $(APP_NAME)..."
	@$(BIN)

# ===============================
#   CODE QUALITY
# ===============================

.PHONY: fmt
fmt: ## Форматує код
	@echo "✨ Форматування коду..."
	go fmt ./...

.PHONY: vet
vet: ## Аналізує код на помилки
	@echo "🔍 Перевірка vet..."
	go vet ./...

.PHONY: lint
lint: ## Лінтинг (golangci-lint)
	@echo "🧠 Лінтинг..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint не знайдено. Встановіть: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: check
check: fmt vet lint ## Повна перевірка якості коду

# ===============================
#   TESTS
# ===============================

.PHONY: test
test: ## Запускає всі unit-тести
	@echo "🧪 Запуск unit-тестів..."
	go test ./... -v -count=1

.PHONY: cover
cover: ## Генерує звіт покриття тестами
	@echo "📊 Генерація звіту покриття..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

# ===============================
#   UTILITIES
# ===============================

.PHONY: deps
deps: ## Оновлює залежності
	@echo "📦 Оновлення залежностей..."
	go mod tidy

.PHONY: db-reset
db-reset: ## Очищає базу даних SQLite
	@echo "🗑️  Видалення локальної бази даних..."
	rm -f $(DB_FILE)
	@echo "✅ osbb.db видалено!"

.PHONY: help
help: ## Показує всі доступні команди
	@echo ""
	@echo "🧭 Доступні команди для проєкту $(APP_NAME):"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""	