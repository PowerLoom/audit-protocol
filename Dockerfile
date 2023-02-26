FROM golang:alpine3.17

ENV GO111MODULE=on

# Install python/pip
ENV PYTHONUNBUFFERED=1
RUN apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools

RUN apk update && apk add --no-cache ethtool nodejs npm bash gcc musl-dev libc-dev python3-dev curl libffi-dev vim nano

RUN npm install pm2 -g
RUN pm2 install pm2-logrotate && pm2 set pm2-logrotate:compress true && pm2 set pm2-logrotate:retain 7

WORKDIR /src
COPY go/go.mod go/go.sum ./
RUN go mod download
# Copy the application's dependencies files
COPY requirements.txt .
RUN pip3 install -r requirements.txt

EXPOSE 9000
EXPOSE 9002
EXPOSE 9030

COPY . .
RUN ./build.sh

RUN chmod +x init_processes.sh snapshotter_autofill.sh
