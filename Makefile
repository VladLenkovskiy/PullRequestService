
build:
	go build ./...

docker-compose-up:
	docker compose up --build

docker-down:
	docker compose down

test-integration:
	go test .\tests\integration\ -v

test-e2e:
	go test .\tests\e2e\ -v

test-load:
	powershell -Command "(Get-Content .env) -replace '^LOAD_MODE=.*', 'LOAD_MODE=load' | Set-Content .env"
	go run .\load\load.go
	powershell -Command "(Get-Content .env) -replace '^LOAD_MODE=.*', 'LOAD_MODE=test' | Set-Content .env"
	go run .\load\load.go

lint:
	golangci-lint run