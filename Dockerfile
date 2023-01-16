FROM ubuntu:22.04

# Update the package list and install the required dependencies
RUN apt-get update -y && apt-get install -y python3-pip git curl golang-go nodejs npm python-is-python3

# Copy the application's dependencies files
COPY requirements.txt .

RUN npm install pm2 -g

# Install the Python dependencies
RUN pip3 install -r requirements.txt

# Set the environment variable for Go modules
ENV GO111MODULE=on

# Set the working directory to the Go source code
WORKDIR /go/src/github.com/powerloom/goutils

# Copy the Go module files and download the dependencies
COPY goutils/go.mod goutils/go.sum ./
RUN go mod download
COPY go-pruning-archival-service/go.mod go-pruning-archival-service/go.sum ./
RUN go mod download
COPY go-payload-commit-service/go.mod go-payload-commit-service/go.sum ./
RUN go mod download
COPY dag_verifier/go.mod dag_verifier/go.sum ./
RUN go mod download
COPY token-aggregator/go.mod token-aggregator/go.sum ./
RUN go mod download

# Install supervisor
RUN pip3 install supervisor

# Copy the rest of the application's files
COPY . .

# Use environment variables for ports
ENV PORT_1 9000
ENV PORT_2 9002
ENV PORT_3 9030

# Expose the ports that the application will listen on
EXPOSE $PORT_1 $PORT_2 $PORT_3

COPY docker/ma.conf /etc/supervisord.conf
RUN chmod +x docker/run_services.sh snapshotter_autofill.sh

