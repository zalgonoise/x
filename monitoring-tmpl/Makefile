.PHONY : default get-loki-driver build clean up down run fresh

default: clean up run

get-loki-driver:
	docker plugin install grafana/loki-docker-driver:2.8.2 --alias loki --grant-all-permissions

build:
	docker compose build

clean:
	docker compose down
	docker compose -f docker-compose.monitor.yaml down
	docker volume rm monitoring-tmpl_grafana monitoring-tmpl_loki monitoring-tmpl_minio monitoring-tmpl_prometheus monitoring-tmpl_tempo

up:
	docker compose -f docker-compose.monitor.yaml up -d

down:
	docker compose down
	docker compose -f docker-compose.monitor.yaml down

run:
	docker compose up

fresh: clean build up run