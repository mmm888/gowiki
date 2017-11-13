From golang:1.9
MAINTAINER mmm888

RUN apt-get update; apt-get -y install vim; git clone https://github.com/mmm888/gowiki ${GOPATH}/src/github.com/mmm888/gowiki; git clone https://github.com/mmm888/wikitest ${GOPATH}/src/github.com/mmm888/gowiki/wikitest

# localhost:8080 を変更しないといけないため
ADD config.toml ${GOPATH}/src/github.com/mmm888/gowiki/config.toml
ADD wikitest ${GOPATH}/src/github.com/mmm888/gowiki/wikitest

WORKDIR ${GOPATH}/src/github.com/mmm888/gowiki
# gowiki 以下の wikitest ディレクトリと config.toml を参照するため、go install . だとエラー
# wikitest の場所を指定できるようにする + wikitest の名前を変更する
RUN go get -u github.com/golang/dep/cmd/dep; dep ensure; go build .

EXPOSE 8080:8080
#VOLUME wikitest:${GOPATH}/src/github.com/mmm888/gowiki/wikitest
#WORKDIR ${GOPATH}

CMD ./gowiki
