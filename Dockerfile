FROM alpine:latest

RUN mkdir /app

COPY build/package/ephemeral-roles /app

RUN chmod +x /app/ephemeral-roles

EXPOSE 5000

CMD /app/ephemeral-roles