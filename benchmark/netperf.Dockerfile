FROM centos:7

LABEL maintainer="bibaijin"

RUN yum install -y gcc
RUN yum install -y make
COPY netperf-2.7.0 /netperf-2.7.0
RUN cd /netperf-2.7.0 && ./configure && make && make install

RUN yum install -y telnet
RUN yum install -y nmap

EXPOSE 12865

COPY run.sh /run.sh

CMD /run.sh