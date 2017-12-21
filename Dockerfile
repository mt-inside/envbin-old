FROM alpine:3.7

ADD . /envbin

CMD cd /envbin && exec ./envbin
