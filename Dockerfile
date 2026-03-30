FROM docker.io/library/alpine:3.23 AS runtime

RUN \
  apk add --update --no-cache \
    bash \
    coreutils \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["satisfaction-survey"]
COPY satisfaction-survey /usr/bin/
COPY static /app/static
WORKDIR /app

EXPOSE 8080

USER 65536:0
