FROM golang:1.15.0-alpine3.12

COPY ephemeral-roles-debug .

RUN apk add --no-cache git libc6-compat
RUN go get -u github.com/go-delve/delve/cmd/dlv

EXPOSE 8080 2345

ENTRYPOINT ["dlv", "--listen=:2345", "--headless=true", "--api-version=2", "exec", "./ephemeral-roles-debug"]
