FROM bibaijin/nginx:1.12.0

LABEL maintainer="bibaijin"

RUN apt-get install -y procps

COPY nginx.conf /usr/local/nginx/nginx.conf