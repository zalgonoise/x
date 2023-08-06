.PHONY : default build clean up run fresh

default: run

build:
	docker compose build

clean:
	docker compose down
	docker compose -f docker-compose.monitor.yaml down
	rm -rf ./local
	mkdir -p ./local/{tempo/data,minio/data,prometheus/data,grafana/{data,provisioning/datasources}}

up:
	docker compose -f docker-compose.monitor.yaml up -d

run:
	docker compose up

fresh: clean build up run