.PHONY: dev prod

dev:
	docker compose -f ./docker/docker-compose.dev.yml -p yamb-dev up --build

prod:
	docker compose -f ./docker/docker-compose.yml -p yamb-prod up -d --build