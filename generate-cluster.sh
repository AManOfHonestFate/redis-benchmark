#!/bin/bash
set -e

# Load .env
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
else
  echo ".env file not found!"
  exit 1
fi

NODES=${REDIS_NODES:-6}
PASSWORD=${REDIS_PASSWORD:-changeme}
TLS=${REDIS_TLS:-no}

# Generate docker-compose.generated.yml
echo "version: '3.8'" >docker-compose.generated.yml
echo "services:" >>docker-compose.generated.yml

for i in $(seq 1 $NODES); do
  cat <<EOF >>docker-compose.generated.yml
  redis-node-$i:
    build: .
    container_name: redis-node-$i
    ports:
      - "$((7000 + i)):6379"
    volumes:
      - ./data/node-$i:/data
    environment:
      - REDIS_PASSWORD=$PASSWORD
    networks:
      - redis-cluster
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "$PASSWORD", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          cpus: '0.50'
          memory: 512M
EOF
done

cat <<EOF >>docker-compose.generated.yml
  redis-exporter:
    image: oliver006/redis_exporter
    container_name: redis-exporter
    ports:
      - "9121:9121"
    environment:
      - REDIS_PASSWORD=$PASSWORD
    networks:
      - redis-cluster

networks:
  redis-cluster:
EOF

# Generate init-cluster.sh
echo "#!/bin/bash" >init-cluster.sh
echo "set -e" >>init-cluster.sh
echo "NODE_IPS=\""$(for i in $(seq 1 $NODES); do echo -n "redis-node-$i:6379 "; done)"\"" >>init-cluster.sh
echo "REPLICAS=1" >>init-cluster.sh
echo "PASSWORD=\"$PASSWORD\"" >>init-cluster.sh
echo 'echo ">>> Creating cluster with nodes: $NODE_IPS"' >>init-cluster.sh
echo 'echo ">>> Number of replicas: $REPLICAS"' >>init-cluster.sh
echo 'docker exec -it redis-node-1 redis-cli -a "$PASSWORD" --cluster create $NODE_IPS --cluster-replicas $REPLICAS --cluster-yes' >>init-cluster.sh
chmod +x init-cluster.sh

