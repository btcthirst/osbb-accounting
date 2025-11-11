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