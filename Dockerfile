FROM redis:7.2

# Uncomment to add TLS support
# COPY certs /certs
# RUN chown -R redis:redis /certs

COPY redis.conf /usr/local/etc/redis/redis.conf
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
