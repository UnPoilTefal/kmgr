FROM gcr.io/distroless/static-debian12

# Le contexte prepare par « dockers_v2 » range les binaires par plateforme
# (linux/amd64/kmgr, linux/arm64/kmgr) : une seule construction sert les deux
# architectures, a condition de passer par $TARGETPLATFORM.
ARG TARGETPLATFORM

COPY $TARGETPLATFORM/kmgr /usr/local/bin/kmgr

ENTRYPOINT ["/usr/local/bin/kmgr"]
