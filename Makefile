run-db:
	docker-compose up -d

stop-db:
	docker-compose down

db-shell:
	docker exec -it titan_postgres psql -U titan_user -d titan_ledger

test:
	go test -v ./...

.PHONY: run-db stop-db db-shell test