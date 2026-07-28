GO ?= go
NPM ?= npm

.PHONY: dev api worker web test lint sqlc db-reset db-migration

dev:
	@echo "TODO: wire local Supabase, API, worker, and web startup"

api:
	$(GO) run ./apps/api

worker:
	$(GO) run ./apps/worker

web:
	cd apps/web && $(NPM) run dev

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

sqlc:
	sqlc generate

db-reset:
	supabase db reset

db-migration:
	@if [ -z "$(name)" ]; then echo "usage: make db-migration name=create_table"; exit 1; fi
	supabase migration new $(name)
