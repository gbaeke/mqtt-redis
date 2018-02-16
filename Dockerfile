
FROM golang:alpine

ADD mqtt-redis /

CMD ["/mqtt-redis"]