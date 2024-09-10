# 使用说明

## 如何编译rubik
### 依赖
- golang > 1.13

### 编译
1. 下载代码
    ```bash
    cd /tmp
    git clone https://gitee.com/openeuler/rubik.git
    ```
2. 进入rubik目录编译rubik二进制
    ```bash
    cd /tmp/rubik
    make # 编译结束后，rubik目录下出现build目录。
    # 安装rubik至/var/lib/rubik目录
    make install
    ```
    install命令将会新建`/var/lib/rubik`目录，并在其下提供rubik二进制和rubik-daemonset.yaml配置文件。

## 如何使用rubik

目前，rubik仅支持以daemonset形式运行在kubernetes集群中。**我们建议在1.17版本以上的kubernetes稳定版本使用rubik**。
我们提供了rubik-daemonset.yaml脚本供用户使用。在执行`make install`命令后，该文件及rubik的二进制文件将会被安装到`/var/lib/rubik`目录下。

1. 构建镜像
    由于运行daemonset依赖于rubik镜像，因此首先需要构建rubik镜像。我们提供了构建docker镜像的方法：
    ```bash
    cd /tmp/rubik
    make image
    ```
    通过上述命令即可构建出rubik镜像，该镜像以`rubik:VERSION-RELEASE`为名。
    用户也可以手动编译，以编译出符合要求的rubik镜像，例如：
    ```bash
    cd /tmp/rubik
    ## 使用我们提供的DockerFile或者自己编写dockerFile
    docker build -f Dockerfile -t rubik:latest .
    ```
    DockerFile模板：
    ```dockerfile
    FROM scratch
    COPY ./rubik /rubik
    ENTRYPOINT ["/rubik"]
    ```
    随后可以通过`docker images| grep rubik`查看到该镜像，镜像名为`rubik:latest`。
    ```bash
    [root@localhost rubik]# docker images| grep rubik
    rubik                                           latest              712d387a34ec        About a minute ago   39.9MB
    ```
2. 修改rubik配置文件rubik-daemonset.yaml。
- 更改镜像名。替换image相关参数`rubik_image_name_and_tag`为上一步编译镜像时使用的镜像名，如rubik:0.0.1-1或者rubik:latest。

  使用sed命令修改：
  ```bash
  sed -i "/image:/s/:.*/: rubik:latest/" rubik-daemonset.yaml
  ```
  或者手动修改：
  ```yaml
  # image: hub.oepkgs.net/cloudnative/rubik:latest
  # 修改rubik镜像
  image: rubik:latest
  ```
- 按需配置特性
rubik提供了多种特性供用户选择。在kubernetes场景下，rubik的config.json配置文件是以ConfigMap形式配置到rubik容器中的。因此用户可以手动修改config.json以获取rubik提供的不同能力。
默认配置如下：
  ```json
  {
    "agent": {
        "logDriver": "stdio",
        "logDir": "/var/log/rubik",
        "logSize": 1024,
        "logLevel": "info",
        "cgroupRoot": "/sys/fs/cgroup",
        "enabledFeatures": [
          "preemption"
        ]
    },
    "preemption": {
        "resource": [
          "cpu",
          "memory"
        ]
    }
  }
  ```
  该配置默认使能了绝对抢占特性（内存与CPU）。如何修改配置可以参考[Rubik配置说明](./config.md)。rubik支持特性可以参考[特性介绍](./feature.md)。为保障rubik正常运行，我们对rubik运行时进行约束，详见[约束限制](./limitation.md)。

3. 以daemonset形式运行rubik镜像。
    ```bash
    kubectl apply -f rubik-daemonset.yaml
    ```
    即可正常运行rubik。
