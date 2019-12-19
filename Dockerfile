FROM alpine:latest

COPY build/package/ephemeral-roles .

EXPOSE 8080

ENTRYPOINT ["./ephemeral-roles"]
