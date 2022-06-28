# 约束限制

- rubik接受HTTP请求并发量上限1000QPS，并发量超过上限则报错。

- rubik接受的单个请求中pod上限为100个，pod数量越界则报错。

- 每个k8s节点只能部署一个rubik，多个rubik会冲突。

- rubik不提供端口访问，只能通过sock通信。

- rubik只接收合法http请求路径及网络协议：http://localhost/（POST）、http://localhost/ping（GET）、http://localhost/version（GET）。

- rubik磁盘使用需求，配额1GB+。

- rubik内存使用需求，配额100MB+。

- 禁止低优先级往高优先级切换。如业务A先被设置为低优先级（-1），接着请求设置为高优先级（0），rubik报错。

- 容器挂载目录时，rubik本地套接字/run/rubik的目录权限需由业务侧保证最小权限（如700）。

- rubik服务端可用时，单个请求超时时间为120s。如果rubik进程进入T、D状态，则服务端不可用，此时服务不会响应任何请求。为了避免此情况的发生，请在客户端设置超时时间，避免无限等待。

- cpu和memory的在线、离线配置需要统一，否则可能导致两个子系统的QoS冲突。

- 使用混部后，原始的cpu share功能存在限制。具体表现为：

  若当前cpu中同时存在在线任务和离线任务，则离线任务的cpu share无法生效。

  若当前cpu中只有在线任务或只有离线任务，cpu share能生效。

- 用户态的优先级反转、smt、cache、numa负载均衡、离线任务的负载均衡，当前不支持。