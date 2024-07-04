include .env

check:
	@echo ${DATABASE_URI}

migrate-create:
	@(printf "Enter migrate name: "; read arg; migrate create -ext sql -dir db/migrations -seq $$arg);

migrate-up:
	migrate -database ${DATABASE_URI} -path ./db/migrations up

migrate-down:
	migrate -database ${DATABASE_URI} -path ./db/migrations down 1

migrate-down-all:
	migrate -database ${DATABASE_URI} -path ./db/migrations down -all

test-cover:
	go test -v -coverprofile=coverage.out ./internal/* && go tool cover -html=coverage.out -o coverage.html

build-mocks:
	@go get github.com/vektra/mockery/v2@v2.43.2
	@~/go/bin/mockery