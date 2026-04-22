DockerfileFROM gcr.io/distroless/static-debian12

COPY kmgr /usr/local/bin/kmgr   # GoReleaser copie le binaire ici

ENTRYPOINT ["/usr/local/bin/kmgr"]
