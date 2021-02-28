FROM golang:1.15-alpine AS builder

# install os deps
RUN apk update && \
    apk upgrade && \
    apk add --no-cache \
        make

WORKDIR /external


# install project dep
ADD go.mod .
ADD go.sum .
RUN export GO111MODULE=on && \
    go mod download

COPY . .

# build external application
RUN make build-external

FROM alpine:3.9

WORKDIR /external
COPY --from=builder /external/bin/external .

EXPOSE 4200

CMD ["./external"]
