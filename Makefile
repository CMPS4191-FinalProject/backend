## run/api: run the cmd/api application

include .envrc
.PHONY: run/api run/api/win db/migrations/up
run/api:
	@bash -c 'set -a && source .envrc && set +a && go run ./cmd/api'
run/api/win:
	@powershell -Command "Get-Content .envrc | ForEach-Object { if ($$_ -match '^([^=]+)=(.*)$$') { $$value = $$matches[2] -replace '^\"(.*)\"$$', '$$1'; [System.Environment]::SetEnvironmentVariable($$matches[1], $$value, 'Process') } }; go run ./cmd/api"
db/migrations/up:
	@migrate -path ./migrations -database ${DB_DSN} up