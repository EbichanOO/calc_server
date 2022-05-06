FROM golang:latest

ENV TZ Asia/Tokyo
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime &&\
    echo $TZ > /etc/timezone
ENV LANG ja_JP.UTF-8

RUN apt update -y &&\
    apt upgrade -y &&\
    apt install -y \
        mecab \
        libmecab-dev \
        mecab-ipadic-utf8

ENV CGO_LDFLAGS="`/usr/bin/mecab-config --libs`"
ENV CGO_CFLAGS="-I`/usr/bin/mecab-config --inc-dir`"

COPY . /go/src/calc_server/
WORKDIR /go/src/calc_server/go/crawler/

CMD ["go", "run", "."]