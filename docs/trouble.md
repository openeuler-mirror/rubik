# 问题集锦
我们在此为您提供常见的rubik使用问题以及解决方法：
- [运行时可能遇到的问题](#运行问题)
- [编译rubik可能遇到的问题](#编译问题)

## 运行问题
---
**问题描述**：
在单节点kubernetes节点中部署rubik失败，通过`kubectl get pod -A -o wide`命令无法查看到任何rubik pod信息。

**可能原因**：
kubernetes默认对主节点增加了污点属性，导致rubik无法被调用master节点上。通过命令`kubectl get no -o yaml | grep taint -A 5`查看master的污点策略。如下：
```yaml
taints:
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
```

**解决方案**：
对`rubik-daemonset.yaml`文件尾部增加容忍度策略，使之能够被调用到master节点上。
```yaml
  #...
  spec:
    ...
    container:
      ...
    volumes:
      ...
    tolerations:
      - key: node-role.kubernetes.io/master
        operator: "Exists"
        effect: "NoSchedule"
```
或者直接删除主节点污点策略。
```bash
kubectl taint nodes master node-role.kubernetes.io/master-
```

---

## 编译问题

**问题描述**：执行make命令失败，报错为：
```bash
$ make
/usr/lib/golang/pkg/tool/linux_amd64/link: running gcc failed: exit status 1
/usr/bin/ld: cannot find -lc: No such file or directory
make: *** [Makefile:61: release] Error 2
```
**可能原因**：
当前linux环境不支持静态编译。某些linux发行版的glibc未安装了静态库libc.a而仅仅安装了动态库lib.so。

**解决方案**：
1. （推荐）安装glibc-static包，获取libc静态库。
2. 尝试动态编译方式，即删除Makefile中`-extldflags=-static`编译选项。但动态编译方式可能存在问题。例如构建rubik镜像的初始镜像无法提供相应依赖库，将导致rubik二进制无法运行。

