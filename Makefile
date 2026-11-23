# Local development. Needs Docker for PostgreSQL.
include dev.env
export

.PHONY: db generate migrate seed serve check test

db:
	docker compose up -d

generate:
	go tool templ generate
	go tool genx

check:
	go tool templ generate
	go tool genx check
	go vet ./...

migrate:
	go run . migrate

seed:
	go run . seed

serve:
	go run . serve

test:
	go test ./...
