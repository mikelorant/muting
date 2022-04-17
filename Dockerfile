FROM golang:1.18-alpine AS build

ENV GOOS=linux
ENV CGO_ENABLED=0

WORKDIR /root/go/src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . ./

RUN mkdir bin

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o bin ./...

FROM alpine:3.15

RUN addgroup -g 10001 muting && \
    adduser -S -D -H -h / -s /usr/sbin/nologin -u 10001 muting && \
    adduser muting muting

COPY --from=build /root/go/src/bin/* /usr/local/bin

USER muting

ENTRYPOINT ["/usr/local/bin/muting"]
CMD ["server"]
