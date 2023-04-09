DB_URL=postgresql://root:entriesMicroServiceDB@entriesmicroservicedb.cviqqzopm7zr.us-east-2.rds.amazonaws.com:5432/entriesMicroServiceDB

newPostgres:
	docker run --name postgresEntries -p 5433:5433 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=entriesMicroServiceDB -d postgres:latest

postgres:
	docker start postgresEntries

createdb:
	docker exec -it postgresEntries createdb --username=root --owner=root entriesMicroServiceDB

dropdb:
	docker exec -it postgresEntries dropdb entriesMicroServiceDB

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go entriesMicroService/db/sqlc Store

.PHONY: network newPostgres postgres createdb dropdb migrateup migratedown sqlc server mock