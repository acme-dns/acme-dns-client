ARG PLATFORM
FROM certbot/certbot:${PLATFORM}-latest

RUN mkdir /etc/acmedns/
COPY renew /etc/periodic/daily
COPY docker-entrypoint.sh /usr/local/bin/
COPY acme-dns-client /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh /etc/periodic/daily/renew

ENTRYPOINT ["docker-entrypoint.sh"]
VOLUME /etc/acmedns/
VOLUME /etc/letsencrypt/
