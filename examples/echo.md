# netserver 与 netperf

## netserver

```
appname: netperf-service

build:
  base: registry.yxapp.xyz/centos:1.0.1
  prepare:
    version: 20170422
    script:
      - cd netperf-2.7.0 && ./configure && make && make install
  script:
    - echo "Success"

service.netserver:
  type: worker
  cmd: netserver -p 8080 -D
  port: 8080
  memory: 64M
  portal:
    allow_clients: "**"
    image: registry.yxapp.xyz/tcp-reverse-proxy:release-1492757416-2266c14a246498038b30f1f53e17b2668f1ff840
    cmd: /lain/app/tcp-reverse-proxy -port 8080 -serviceproctype worker -serviceprocname netserver
    port: 8080
```

> - netperf-service 的源代码在 [https://github.com/bibaijin/netperf-service](https://github.com/bibaijin/netperf-service)。
> - portal 部分用到了 tcp-reverse-proxy:
>     - image 使用了打包好的 tcp-reverse-proxy image
>     - cmd:
>         - port 默认为 8080
>         - serviceproctype 默认为 worker
>         - serviceprocname 默认为 ${service_name}，在本例中为 `service.netserver` 中的 `netserver`

## netperf

```
appname: netperf-client

build:
  base: registry.yxapp.xyz/centos:1.0.1
  prepare:
    version: 20170422
    script:
      - cd netperf-2.7.0 && ./configure && make && make install
  script:
    - cp -f entry.sh /entry.sh

use_services:
  netperf-service:
    - netserver

proc.worker:
  cmd: /entry.sh
  volumes:
    - /lain/app/benchmark
  memory: 64M
```

> - netperf-client 的源代码在 [https://github.com/bibaijin/netperf-client](https://github.com/bibaijin/netperf-client)。
