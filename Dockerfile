FROM ubuntu:latest
RUN apt-get update -y && apt-get install -y python3-pip python-dev git
EXPOSE 9000
EXPOSE 9002
COPY . .
RUN pip3 install -r requirements.txt
RUN pip3 install supervisor
COPY docker/ma.conf /etc/supervisord.conf
CMD chmod +x docker/run_services.sh
CMD ./run_services.sh
