FROM alpine:latest

COPY promi /usr/bin/promi
COPY packaging/docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod 750 /docker-entrypoint.sh

CMD ["server"]
ENTRYPOINT ["/docker-entrypoint.sh"]
