FROM golang:1.8

LABEL maintainer="bibaijin (bibaijin@gmail.com)"

RUN apt-get update
RUN apt-get install -y net-tools

WORKDIR $GOPATH/src/github.com/laincloud/proxyd

COPY benchmark/netperf-2.7.0 benchmark/netperf-2.7.0
RUN cd benchmark/netperf-2.7.0 && ./configure && make && make install

COPY Godeps Godeps
COPY vendor vendor
COPY *.go ./
COPY log log
RUN go install
RUN ln -s $GOPATH/bin/proxyd /proxyd

COPY test.sh /test.sh

CMD /proxyd