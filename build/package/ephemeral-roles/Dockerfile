FROM gcr.io/distroless/static as base
FROM scratch

COPY --from=base /etc/ssl/certs /etc/ssl/certs
COPY --from=base /usr/share/zoneinfo /usr/share/zoneinfo
COPY ephemeral-roles .

EXPOSE 8081

ENTRYPOINT ["./ephemeral-roles"]
