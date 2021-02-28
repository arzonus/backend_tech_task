FROM golang:1.15-alpine AS builder

# install os deps
RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
        make

WORKDIR /api


# install project dep
ADD go.mod .
ADD go.sum .
RUN export GO111MODULE=on && \
    go mod download

COPY . .

# build api application
RUN make build-api

FROM alpine:3.9

WORKDIR /api
COPY --from=builder /api/bin/api .

ENV GRPC_ADDRESS=":8080"
EXPOSE 8080

CMD ["./api"]
