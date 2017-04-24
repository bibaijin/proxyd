# echo-service 与 echo-client

## [echo-service](https://github.com/bibaijin/echo-service)

```
appname: echo-service

build:
  base: golang:1.8
  prepare:
    version: 201704220021
  script:
    - mkdir -p $GOPATH/src/github.com/bibaijin/echo-service/
    - cp -rf . $GOPATH/src/github.com/bibaijin/echo-service/
    - cd $GOPATH/src/github.com/bibaijin/echo-service/ && go install

release:
  dest_base: registry.yxapp.xyz/centos:1.0.1
  copy:
    - src: $GOPATH/bin/echo-service
      dest: /echod

service.echod:
  type: worker
  cmd: /echod
  port: 8080
  portal:
    allow_clients: "**"
    image: bibaijin/proxyd:1.0.0
    cmd: /proxyd -port 8080 -serviceproctype worker -servicename echod
    port: 8080
```

> - 其中，portal 用到了 proxyd:
>     - image 使用了上传到 docker hub 的 proxyd 镜像
>     - cmd:
>         - port 默认为 8080
>         - serviceproctype 默认为 worker
>         - servicename 为 ${service_name}，在本例中为 `service.netserver` 中的 `netserver`

## [echo-client](https://github.com/bibaijin/echo-client)

```
appname: echo-client

build:
  base: golang:1.8
  prepare:
    version: 201704220054
  script:
    - mkdir -p $GOPATH/src/github.com/bibaijin/echo-client/
    - cp -rf . $GOPATH/src/github.com/bibaijin/echo-client/
    - cd $GOPATH/src/github.com/bibaijin/echo-client/ && go install

release:
  dest_base: registry.yxapp.xyz/centos:1.0.1
  copy:
    - src: $GOPATH/bin/echo-client
      dest: /echo

use_services:
  echo-service:
    - echod

proc.worker:
  cmd: /echo
```
