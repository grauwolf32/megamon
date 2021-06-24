FROM ubuntu:latest

RUN mkdir -p /app/web/frontend/ && mkdir /app/config/ && mkdir /app/files/
COPY ./web/frontend/ /app/web/frontend/
COPY ./config/ /app/config/
COPY ./megamon /app/
WORKDIR "/app/"
ENTRYPOINT "/app/megamon"
