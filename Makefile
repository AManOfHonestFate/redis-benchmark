.PHONY: all
all: build

.PHONY: up
up:
	docker-compose -f docker-compose.generated.yml up -d

.PHONY: down
down:
	docker-compose -f docker-compose.generated.yml down

.PHONY: clean
clean:
	docker-compose -f docker-compose.generated.yml down -v --remove-orphans
	rm -f docker-compose.generated.yml
	rm -f init-cluster.sh
	rm -rf data

.PHONY: backup
backup:
	./backup.sh

.PHONY: restore
restore:
	./restore.sh redis-backup.tar.gz

.PHONY: init-cluster
init-cluster:
	./init-cluster.sh

.PHONY: generate-cluster
generate-cluster:
	./generate-cluster.sh
