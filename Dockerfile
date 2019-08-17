FROM alpine:latest

MAINTAINER yuli@yulibaozi.com

ADD ./go /go

WORKDIR /go/bin

RUN chmod 777 /go/bin/injection

CMD ["/go/bin/injection"]
