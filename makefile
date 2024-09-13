ENVFILE=.env.template

ifneq ("$(wildcard $(ENVFILE))","")
	include $(ENVFILE)
	export $(shell sed 's/=.*//' $(ENVFILE))
endif

# create local db for testing
DB_CONTAINER_NAME=avi-db
create-mock-db:
	docker run --name avi-db \
		-p $(POSTGRES_PORT):5432 \
		-e POSTGRES_USER=$(POSTGRES_USERNAME) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DATABASE) \
		-d postgres

psql-mock-db:
	docker exec -i $(DB_CONTAINER_NAME) psql -U $(POSTGRES_USERNAME) $(POSTGRES_DATABASE)

init-mock-db:
	cat ./init-mock-db.sql | make psql-mock-db

create-init-mock-db:
	make create-mock-db
	sleep 3 # docker container needs more time to be ready "&&" won't help
	make init-mock-db

stop-mock-db:
	docker container stop $(DB_CONTAINER_NAME)

rm-mock-db:
	docker container rm $(DB_CONTAINER_NAME)

srm-mock-db:
	make stop-mock-db && make rm-mock-db

run:
	go run ./cmd

build:
	docker build -t avi .

run-c:
	docker run --name avi --env-file .env -p 8080:8080 avi

rm-c:
	docker container rm avi

logs-c:
	docker logs avi

lint:
	golines -w .
	gofmt -w .

check:
	golangci-lint run 