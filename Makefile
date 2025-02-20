DB_URL=sqlite3://./lms.db

.PHONY: new_migration
.PHONY: migrateup
.PHONY: migratedown
.PHONY: swagger
.PHONY: start_cache

new_migration: 
	migrate create -ext sql -dir db/migration -seq $(name)

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

start_cache:
	docker run -p 6378:6379 --ulimit memlock=-1 docker.dragonflydb.io/dragonflydb/dragonfly

swagger:
	swag init -g main.go -o docs