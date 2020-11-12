FROM alpine:latest as base

RUN apk --no-cache add tzdata

FROM scratch

COPY --from=base /etc/ssl/certs /etc/ssl/certs
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY ephemeral-roles .

EXPOSE 8081

ENTRYPOINT ["./ephemeral-roles"]
