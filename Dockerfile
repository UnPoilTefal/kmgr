FROM gcr.io/distroless/static-debian12

COPY kmgr /usr/local/bin/kmgr

ENTRYPOINT ["/usr/local/bin/kmgr"]
