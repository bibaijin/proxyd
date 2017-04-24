# 性能测试

proxyd 使用 [netperf](http://www.netperf.org/netperf/) 测试。

## 环境

- 1 台 2 核 CPU 和 2 G 内存的虚拟机，操作系统为 Arch Linux
- 虚拟机所在的物理机为 MacBook Pro (Retina, 13-inch, Late 2013)

## 步骤

```
cd ${proxyd-project}
docker-compose build
docker-compose run --rm benchmark sleep 1s  # 启动依赖服务
docker-compose run --rm benchmark > stats.txt  # 测试
```

## 说明

- 测试数据在 [stats.txt](stats.txt)
