# docker run --rm -it pipo bash
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

ENV CGO_LDFLAGS="-L/usr/lib/x86_64-linux-gnu -lmecab -lstdc++"
ENV CGO_CFLAGS="-I/usr/include"

COPY . /go/src/calc_server/
WORKDIR /go/src/calc_server/go/crawler/

CMD ["go", "run", "."]