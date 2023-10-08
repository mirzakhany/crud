.PHONY: up
up:
	dbctl start pg -m ./migrations -f ./test_data -d

.PHONY: down
down:
	dbctl stop pg

.PHONY: run
run:
	go run server.go
