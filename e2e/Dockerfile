FROM golang:1.13

RUN apt-get update && apt install wait-for-it

WORKDIR /app

ADD go.mod go.sum ./
RUN go mod download

ADD common ./common
ADD e2e ./e2e

WORKDIR /app/e2e

ENTRYPOINT ["wait-for-it", "api:80", "-t", "30", "--", "go", "test", "-timeout", "30s"]
