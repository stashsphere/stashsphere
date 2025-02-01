migrate-dev: migrations
	migrate -source file://./migrations -database "postgres://stashsphere:secret@127.0.0.1:5432/stashsphere?sslmode=disable" up

migrate-drop:
	migrate -source file://./migrations -database "postgres://stashsphere:secret@127.0.0.1:5432/stashsphere?sslmode=disable" drop
	
boil:
	sqlboiler psql
