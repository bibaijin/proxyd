# proxyd

一个透明的 tcp 反向代理，用于 [LAIN](https://github.com/laincloud/lain)
[service](https://laincloud.gitbooks.io/white-paper/usermanual/service.html)。

## 示例

- [echo](examples/echo.md)

## 性能测试

条件 \ 指标 | TCP_RR | TCP_STREAM | TCP_MAERTS | TCP_SENDFILE
----------- | ------ | ---------- | ---------- | ------------
没有 proxy | 15109 trans/s | 13919 Mbits/s | 13526 Mbits/s | 13098 Mbits/s
nginx | 7059 trans/s | 10585 Mbits/s | 8897 Mbits/s | 8311 Mbits/s
proxyd | 8541 trans/s | 11393 Mbits/s | 10568 Mbits/s | 12953 Mbits/s

> - TCP_RR 指 TCP Request/Response 测试:
>     - trans/s 指 transaction/s，即每秒完成的交易数，即 qps
> - TCP_STREAM 指由客户端向服务器发送数据的吞吐量测试
> - TCP_MAERTS 指由服务器向客户端发送数据的吞吐量测试（MAERTS 是 STREAM 的逆序）
> - TCP_SENDFILE 与 TCP_STREAM 类似，不过调用系统的 sendfile() 传输数据
>   而非 TCP_STREAM 使用的 send() 
> - 详见 [benchmark/README.md](benchmark/README.md)

## 设计

+ proxyd 从 lainlet 获取 upstream 信息
+ 目前采用 round-robin 方法，将客户端请求轮流连接到不同的 upstream

## 构建及上传镜像

```
lain build  # 构建 proxyd:release 镜像
lain tag ${LAIN-cluster}  # 打标签为 registry.${LAIN-domain}/proxyd:release-${timestamp}-${git-commit-id}
lain push ${LAIN-cluster}  # 上传镜像到 registry.${LAIN-domain}
```
