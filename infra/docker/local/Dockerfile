FROM golang:1.26.2

RUN groupadd -g 1000 app && \
    useradd -g 1000 -u 1000 app && \
    mkdir -p /var/www /home/app/.cache


RUN go install github.com/go-delve/delve/cmd/dlv@master

COPY ./ /var/www/.
WORKDIR /var/www

RUN chown -R 1000:1000 /var/www /home/app /go /tmp && \
    chmod 777 /tmp && \
    chmod +t /tmp
#USER app

EXPOSE 4000 8002
