DB_URL=sqlite3://./lms.db

.PHONY: new_migration
.PHONY: migrateup
.PHONY: migratedown
.PHONY: swagger

new_migration: 
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

swagger:
	swag init -g main.go -o docs
