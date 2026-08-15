# Node runtime on top of the repo's static binary image.
# Build order:
#   docker build -t neutaro:local .            # repo root — the existing Dockerfile
#   docker build -t neutaro:node docker/       # this file
FROM neutaro:local
RUN apk add --no-cache curl jq
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
VOLUME /neutaro
EXPOSE 26656 26657
ENTRYPOINT ["/entrypoint.sh"]
