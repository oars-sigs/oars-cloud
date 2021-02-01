# OarsCloud 使用说明

OarsCloud 目前设计分为集群、服务、网关、配置四大模块，命名空间用于应用命名隔离，后续可用于权限管理。集群模块主要是集群概览、命名空间管理；服务模块用于服务管理；网关是作为入口管理（类k8s 的 ingress ）；配置用于配置管理（目前还不可用）


## 集群

### 集群概览

包括集群资源总览（cpu 内存 节点数）和节点运行状态和系统信息


### 命名空间

可以创建、删除命名空间，名字命名规范为支持数字、小写字母和'-'

## 服务

### 服务管理

可以创建、编辑和删除服务，名字命名规范为支持数字、小写字母和'-'。服务内容使用docker-compose 规范（大部分，非全部），新加模板和configmap


示例1：

```yaml

image: "busybox"
command: 
  - "sh"
  - "-c"
  - "echo {{ .Endpoint.Domain }} && sleep 3600"
configmap: 
  /etc/test/test.conf: |
    hostname: {{ .Endpoint.Hostname }}
volumes:
- /opt/worker/demo/busybox/data:/data
environment:
  - "TEST_VAR=test"
restart: always
privileged: true
resource:
  cpu: 0.1 #0.1核cpu
  memory: 100 #100m 内存

```

- 模板引擎使用的是golang 的template

`{{ .Endpoint.Domain }}` 当前端点的域名

`{{ .Endpoint.Hostname }}` 当前端点的节点主机名

`{{ .Endpoint.Name }}` 当前端点的名称，默认该节点的主机名


- configmap: 配置文件挂载，上面例子是生成test.conf文件，并挂载给容器`/etc/test/test.conf`

- image: 容器镜像，xxx:tag

- volumes: 挂载存储，目前仅支持本机存储，`主机目录:容器目录`。如果主机目录为绝对路径直接使用该路径（不会自动创建目录），如果机目录为相对路径使用worker 目录下创建（默认/opt/oars/worker/volume/{namespace}/{service}/xxxx/xxxx ）

- environment： 环境变量

- restart： 重启策略。（ no，默认策略，在容器退出时不重启容器；
on-failure，在容器非正常退出时（退出状态非0），才会重启容器；
on-failure:3，在容器非正常退出时重启容器，最多重启3次；
always，在容器退出时总是重启容器；
unless-stopped，在容器退出时总是重启容器，但是不考虑在Docker守护进程启动时就已经停止了的容器 ）

- command: 命令行

- privileged: 是否使用特权 

- network_mode：网络模式，host （主机网络）,bridge（桥网络，默认）,none （无网络）

- user：设置用户

- resource： 设置资源限制

### 服务端点

端点管理，可以重启端点，停止端点，查看端点事件、日志和端点命令行工具。（一个端点即一个容器）

- 端点事件： 端点创建、删除、启动事件，如果一个端点一直起不来可以查看事件

- 端点日志：容器日志，仅展示后100行（后续优化）


## 网关

入口管理，换句话说就是将集群内的服务暴露给集群外

### 端口管理

创建一个监听端口,名字命名规范为支持数字、小写字母和'-'。

示例

```yaml

port: 443

```

### 路由配置

为一个端口创建一个路由，支持http 和tcp

http 示例

```yaml
rules: 
- host: "admin.oars.gzsunrun.cn"
  http: 
    paths: 
    - path: "/"
      backend: 
        serviceName: "admin"
        servicePort: 8801
```

- host: 为访问域名，可以留空，使用主机ip 访问

- backend.serviceName 服务名，所关联的服务名称 

- backend.servicePort 服务端口，所关联的服务端口

就可以通过 `https://admin.oars.gzsunrun.cn` 访问服务了

tcp 示例

```yaml
rules: 
- tcp: 
    backend: 
      serviceName: "admin"
      servicePort: 8801
```

注意：

一个端口只能设置`一条tcp路由`或`多条http路由`

## 配置

暂未开放




