

.PHONY: default run stop clean

default:
	docker compose up -d

upgrade:
	docker compose build
	docker compose down
	docker compose up -d
	docker image prune -f

clean:
	docker compose down -v --rmi all
