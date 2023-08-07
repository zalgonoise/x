.PHONY : default build clean up down run fresh

default: clean up run

build:
	docker compose build

clean:
	docker compose down
	docker compose -f docker-compose.monitor.yaml down
	sudo rm -rf ./local
	mkdir -p ./local/{tempo/data,prometheus/data,grafana/data,minio/data}

up:
	docker compose -f docker-compose.monitor.yaml up -d

down:
	docker compose down
	docker compose -f docker-compose.monitor.yaml down

run:
	docker compose up

fresh: clean build up run