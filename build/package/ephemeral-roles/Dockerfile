FROM alpine:latest as base

FROM scratch

COPY --from=base /etc/ssl/certs /etc/ssl/certs
COPY ephemeral-roles .

EXPOSE 8080

ENTRYPOINT ["./ephemeral-roles"]
