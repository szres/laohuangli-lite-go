

.PHONY: default run stop clean

default:
	docker build --rm -t laohuangli:dev .
	docker image prune --filter label=stage=builder -f

run:
	docker compose up -d

stop:
	docker compose down

clean:
	docker compose down -v --rmi all
