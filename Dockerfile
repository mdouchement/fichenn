# build stage
FROM golang:alpine as build-env
LABEL maintainer="mdouchement"

RUN apk upgrade
RUN apk add --update --no-cache git curl

ARG CHECKSUM_VERSION=v0.2.1
ARG CHECKSUM_SUM=b87278322e1bdb080f709bdd14beaad0b92be6879be9edf828f9800bb074ad52

RUN curl -L https://github.com/mdouchement/checksum/releases/download/$CHECKSUM_VERSION/checksum-linux-amd64 -o /usr/local/bin/checksum && \
    echo "$CHECKSUM_SUM  /usr/local/bin/checksum" | sha256sum -c && \
    chmod +x /usr/local/bin/checksum

ARG TASK_VERSION=v3.11.0
ARG TASK_SUM=8284fa89367e0bbb8ba5dcb90baa6826b7669c4a317e5b9a46711f7380075e21

RUN curl -LO https://github.com/go-task/task/releases/download/$TASK_VERSION/task_linux_amd64.tar.gz && \
    checksum --verify=$TASK_SUM task_linux_amd64.tar.gz && \
    tar -xf task_linux_amd64.tar.gz && \
    cp task /usr/local/bin/

RUN mkdir -p /go/src/github.com/mdouchement/fichenn
WORKDIR /go/src/github.com/mdouchement/fichenn

ENV CGO_ENABLED 0
ENV GOPROXY https://proxy.golang.org

COPY . /go/src/github.com/mdouchement/fichenn
# Dependencies
RUN go install github.com/vugu/vugu/cmd/vugugen@latest
RUN go install github.com/vugu/vgrouter/cmd/vgrgen@latest
RUN go mod download

RUN task webfinn-build

# final stage
FROM alpine
LABEL maintainer="mdouchement"

COPY --from=build-env /go/src/github.com/mdouchement/fichenn/bin/webfinn /usr/local/bin/

ENV WEBFINN_CONFIG /etc/webfinn/webfinn.yml

EXPOSE 5000
CMD ["webfinn"]
