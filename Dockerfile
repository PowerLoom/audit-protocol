FROM ubuntu:22.04
RUN apt-get update -y && apt-get install -y python3-pip git curl golang-go nodejs npm
RUN python3 -v
RUN pip3 -v
RUN node -v
RUN go version
RUN npm install pm2 -g
RUN pm2 ls
EXPOSE 9000
EXPOSE 9002
EXPOSE 9030
COPY requirements.txt .
RUN pip3 install -r requirements.txt
ENV GO111MODULE=on
WORKDIR /go/src/github.com/powerloom/goutils
COPY goutils/go.mod .
COPY goutils/go.sum .
RUN go mod download
COPY go-pruning-archival-service/go.mod .
COPY go-pruning-archival-service/go.sum .
RUN go mod download
COPY go-payload-commit-service/go.mod .
COPY go-payload-commit-service/go.sum .
RUN go mod download
COPY dag_verifier/go.mod .
COPY dag_verifier/go.sum .
RUN go mod download
COPY token-aggregator/go.mod .
COPY token-aggregator/go.sum .
RUN go mod download
RUN pip3 install supervisor
RUN apt-get install -y python-is-python3
COPY . .
COPY docker/ma.conf /etc/supervisord.conf
CMD chmod +x docker/run_services.sh
CMD chmod +x snapshotter-autofill.sh
#CMD ./run_services.sh
