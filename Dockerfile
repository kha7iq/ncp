FROM alpine:latest
ENTRYPOINT ["/usr/bin/ncp"]
COPY ncp /usr/bin/ncp