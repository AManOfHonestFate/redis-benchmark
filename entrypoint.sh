#!/bin/sh
set -e

# Template redis.conf with password if set
if [ -n "$REDIS_PASSWORD" ]; then
  cp /usr/local/etc/redis/redis.conf /tmp/redis.conf
  echo "requirepass $REDIS_PASSWORD" >> /tmp/redis.conf
  echo "masterauth $REDIS_PASSWORD" >> /tmp/redis.conf
  exec redis-server /tmp/redis.conf
else
  exec redis-server /usr/local/etc/redis/redis.conf
fi 