

.PHONY: default down upgrade clean

default:
	docker compose up $(service) -d

down:
	docker compose down $(service)

upgrade:
	docker compose build $(service)
	docker compose down $(service)
	docker compose up $(service) -d
	docker image prune -f

clean:
	docker compose down -v --rmi all
