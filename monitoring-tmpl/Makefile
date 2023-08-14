.PHONY : default get-loki-driver build clean up down run fresh

default: new build

get-loki-driver:
	docker plugin install grafana/loki-docker-driver:2.8.2 --alias loki --grant-all-permissions

build:
	docker compose build

clean: down
	docker volume rm monitoring-tmpl_grafana monitoring-tmpl_loki monitoring-tmpl_minio monitoring-tmpl_prometheus monitoring-tmpl_tempo

up:
	docker compose -f docker-compose.monitor.yaml up minio -d
	docker compose -f docker-compose.monitor.yaml up tempo loki -d
	docker compose -f docker-compose.monitor.yaml up prom mimir-1 mimir-2 mimir-lb -d
	docker compose -f docker-compose.monitor.yaml up grafana -d

down:
	docker compose down
	docker compose -f docker-compose.monitor.yaml down grafana
	docker compose -f docker-compose.monitor.yaml down prom
	docker compose -f docker-compose.monitor.yaml down mimir-lb
	docker compose -f docker-compose.monitor.yaml down mimir-1 mimir-2
	docker compose -f docker-compose.monitor.yaml down tempo
	docker compose -f docker-compose.monitor.yaml down loki
	docker compose -f docker-compose.monitor.yaml down minio

restart: down up

new: clean up

run:
	docker compose up

fresh: new build run