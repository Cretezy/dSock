# Development stage with gin and all required files
FROM golang:1.13 AS development

RUN go get -u github.com/cosmtrek/air

WORKDIR /app

ADD go.mod go.sum ./
RUN go mod download

ADD common ./common
ADD api ./api

WORKDIR /app/api
ENV PORT 80
EXPOSE 80

ENTRYPOINT ["air"]


# Release builder stage, to build the output binary
FROM development AS release_builder

COPY --from=development /app /app

ENV CGO_ENABLED=0
RUN go build -o build/app -ldflags "-s -w"


# Release stage, with only the binary
FROM scratch AS release

COPY --from=release_builder /app/api/build/app /app

ENV PORT 80
EXPOSE 80

ENTRYPOINT ["/app"]
