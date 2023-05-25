FROM golang:alpine3.17

ENV GO111MODULE=on

RUN rm -rf /var/cache/apk/* && \
    rm -rf /tmp/*

RUN apk update && apk add --no-cache ethtool nodejs npm bash gcc musl-dev libc-dev curl libffi-dev vim nano ca-certificates

RUN npm install pm2 -g
RUN pm2 install pm2-logrotate && pm2 set pm2-logrotate:compress true && pm2 set pm2-logrotate:retain 7

WORKDIR /src
COPY go/go.mod go/go.sum ./
RUN go mod download

EXPOSE 9000

COPY . .
RUN ./build.sh

RUN chmod +x init_processes.sh snapshotter_autofill.sh
