migrate-dev: migrations
	migrate -source file://./migrations -database "postgres://$(PGUSER):$(PGPASSWORD)@$(PGHOST):$(PGPORT)/$(PGDATABASE)?sslmode=disable" up

migrate-drop:
	migrate -source file://./migrations -database "postgres://$(PGUSER):$(PGPASSWORD)@$(PGHOST):$(PGPORT)/$(PGDATABASE)?sslmode=disable" drop

boil:
	sqlboiler psql

dump-dev:
	pg_dump --clean --if-exists > $(shell date +%Y-%m-%d-%H%M%S)-dump.sql

restore-dev:
	psql < $(filter-out $@,$(MAKECMDGOALS))

backup-store:
	tar -cvf $(shell date +%Y-%m-%d-%H%M%S)-image_store.tar image_store

restore-store:
	rm -rf image_store && tar -xvf $(filter-out $@,$(MAKECMDGOALS))

%:
	@:
