.PHONY: translate setup run_worker run_market run_admin

translate:
	go generate ./internal/translations
	go tool translate -locales-dir=./internal/translations/locales
	go generate ./internal/translations

generate:
	go tool sqlc generate
	go tool templ generate

setup:
	go tool sqlc generate
	go tool templ generate

run_worker:
	$(MAKE) setup
	go run ./cmd/worker --addr=0.0.0.0:4420 --dsn=postgresql://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_NAME)?sslmode=disable --minio-endpoint=localhost:9000 --dev=true

run_market:
	$(MAKE) setup
	go run ./cmd/market --addr=0.0.0.0:4000 --dsn=postgresql://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_NAME)?sslmode=disable --minio-endpoint=localhost:9000 --dev=true --entry-guard=false --captcha=false

run_admin:
	$(MAKE) setup
	go run ./cmd/admin --addr=0.0.0.0:4002 --dsn=postgresql://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_NAME)?sslmode=disable --minio-endpoint=localhost:9000 --dev=true

run_seed:
	$(MAKE) setup
	go run ./cmd/seed --addr=0.0.0.0:4000 --dsn=postgresql://$(DB_USER):$(DB_PASS)@localhost:5432/$(DB_NAME)?sslmode=disable --minio-endpoint=localhost:9000 --dev=true --entry-guard=false

