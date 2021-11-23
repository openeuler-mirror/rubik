# http接口

rubik包含如下http接口

## 设置、更新Pod优先级接口

接口语法：

```bash
HTTP POST /run/rubik/rubik.sock
{
    "Pods": {
        "podaaa": {
            "CgroupPath": "kubepods/burstable/podaaa",
            "QosLevel": 0
        },
        "podbbb": {
            "CgroupPath": "kubepods/burstable/podbbb",
            "QosLevel": -1
        }
    }
}
```

参数说明：

- pods map必须提供pods。

- podUID map必须提供每个pod的UID。

- QosLevel int必须提供优先级。

  - 0：默认值，在线业务。
  - -1：离线业务。
  - 其他：非法，不支持。

- CgroupPath string必须提供Pod的cgroup子路径。

说明：

- 请求并发量1000QPS，并发量越界报错。
- 单个请求pod数100个，请求数量越界报错。

示例如下：

```sh
curl -v -H "Accept: application/json" -H "Content-type: application/json" -X POST --data '{"Pods": {"podaaa": {"CgroupPath": "kubepods/burstable/podaaa","QosLevel": 0},"podbbb": {"CgroupPath": "kubepods/burstable/podbbb","QosLevel": -1}}}' --unix-socket /run/rubik/rubik.sock http://localhost/
```

## 探活接口

rubik作为HTTP服务，提供探活接口用于帮助判断rubik服务是否还在运行。

接口形式：HTTP/GET /ping

示例如下：

```sh
curl -XGET --unix-socket /run/rubik/rubik.sock http://localhost/ping
```

## 版本信息查询接口

rubik支持通过HTTP请求查询版本号。

接口形式：HTTP/GET /version

示例如下：

```sh
curl -XGET --unix-socket /run/rubik/rubik.sock http://localhost/version
{"Version":"0.0.1","Release":"1","Commit":"29910e6","BuildTime":"2021-05-12"}
```

