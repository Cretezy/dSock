FROM golang:1.13

WORKDIR /app

ADD go.mod go.sum ./
RUN go mod download

ADD common ./common
ADD e2e ./e2e

WORKDIR /app/e2e

ENTRYPOINT ["/bin/sh", "-c", "sleep 1 && go test -timeout 10s"]
