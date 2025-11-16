dir := bin
app := osbb-accounting
mask := ./${dir}/${app}

build:
	go build -o ${mask} ./cmd/main.go

run:
	${mask}

example:
	go run ./cmd/example/main.go

new:
	rm -rf ~/.osbb-accounting/
	go build -o ${mask} .
	${mask}

clean:
	go clean -cache

tests:	# Запуск всіх тестів	
	go test ./...

tests-c:	# Тести з покриттям
	go test -cover ./...

tests-p:	# Тести конкретного пакету
	go test ./service

# Міграції
migrate:
	go run ./cmd/migrate/main.go migrate

migrate-rollback:
	go run ./cmd/migrate/main.go rollback

migrate-validate:
	go run ./cmd/migrate/main.go validate

db-version:
	go run ./cmd/migrate/main.go version

# Швидке виправлення проблеми
migrate-repair:
	go run ./cmd/migrate/main.go repair

# Повний скидання БД (ВИДАЛЯЄ ВСІ ДАНІ!)
migrate-reset:
	go run ./cmd/migrate/main.go reset

# Статус міграцій
migrate-status:
	go run ./cmd/migrate/main.go status