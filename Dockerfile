
FROM redis:7.2

# Uncomment to add TLS support
# COPY certs /certs
# RUN chown -R redis:redis /certs

COPY redis.conf /usr/local/etc/redis/redis.conf

ARG REDIS_PASSWORD=changeme
ENV REDIS_PASSWORD=$REDIS_PASSWORD

CMD ["sh", "-c", "exec redis-server /usr/local/etc/redis/redis.conf --requirepass \"$REDIS_PASSWORD\" --masterauth \"$REDIS_PASSWORD\""]
