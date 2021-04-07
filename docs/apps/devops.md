# devops 安装

## gitlab 安装

新建命名空间 `devops`

新建服务 `gitlab`

内容为以下：

```yaml
image: gitlab/gitlab-ce:13.2.9-ce.0
restart: always
environment:
- GITLAB_OMNIBUS_CONFIG=external_url 'http://192.168.1.120:11000'
ports:
- '11000:11000'
- '2222:22'
volumes:
- 'config:/etc/gitlab'
- 'logs:/var/log/gitlab'
- 'data:/var/opt/gitlab'
- '/usr/share/zoneinfo/Asia/Shanghai:/etc/localtime'

```

注意修改 `hostname` 为访问地址ip，external_url 修改为http://ip:11000


 服务起来后，访问http://ip:11000，配置root 密码


 ## mysql 安装

 新建服务 `mysql`

内容

 ```yaml

image: mysql:5.7.30
command:
- --character-set-server=utf8
- --collation-server=utf8_general_ci
environment:
- MYSQL_ROOT_PASSWORD=123456
- MYSQL_DATABASE=drone
- MYSQL_USER=drone
- MYSQL_PASSWORD=123456
restart: always
volumes:
- data:/var/lib/mysql
- /etc/localtime:/etc/localtime


 ```

`MYSQL_ROOT_PASSWORD`: mysql root 密码注意修改

`MYSQL_PASSWORD`:  drone 数据库用户密码


 ## drone 安装 

 1、 申请gitlab oauth2 认证信息

 用户Setting-> Applications

 ![](https://docs.drone.io/screenshots/gitlab_token_create.png)

 ![](https://docs.drone.io/screenshots/gitlab_token_created.png)

 新建服务 `drone` 

 内容

 ```yaml

image: registry.cn-shenzhen.aliyuncs.com/oars/drone
ports:
- '38080:80'
environment:
- DRONE_GIT_ALWAYS_AUTH=false
- DRONE_GITLAB_SERVER=https://git.cloud.gzsunrun.cn 
- DRONE_GITLAB_CLIENT_ID=8071229e164a7d0dfdd6cb0a2d0a1244762bece1f5cdf89e0d36fb9dd08d908e
- DRONE_GITLAB_CLIENT_SECRET=82c91a5cd2c1362ce54bc5a11b3210b74893477136e058a6f7da31b416e47f35
- DRONE_SERVER_HOST=192.168.1.120:38080
- DRONE_SERVER_PORT=:80
- DRONE_SERVER_PROTO=http
- DRONE_RUNNER_CAPACITY=10
- DRONE_TLS_AUTOCERT=false
- DRONE_LOGS_DEBUG=true
- DRONE_DATABASE_DRIVER=mysql
- DRONE_DATABASE_DATASOURCE=drone:123456@tcp(mysql.devops:3306)/drone
- DRONE_RPC_SECRET=qwer123
- DRONE_RUNNER_PRIVILEGED_IMAGES=plugins/docker


 ```

`DRONE_GITLAB_SERVER` ： gitlab 地址

`DRONE_GITLAB_CLIENT_ID` : 刚才申请的gitlab Applications ID

`DRONE_GITLAB_CLIENT_SECRET` :  刚才申请的gitlab Applications secret

`DRONE_SERVER_HOST`： 访问的主机地址，注意改ip

`DRONE_DATABASE_DATASOURCE` : mysql 连接地址，主要修改数据库密码

访问http://192.168.1.120:38080

## drone-runner 安装 

新建服务 `drone-runner`



 ```yaml
image: drone/drone-runner-docker:1
environment:
- DRONE_RPC_PROTO=http
- DRONE_RPC_HOST=192.168.1.120:38080
- DRONE_RPC_SECRET=qwer123
- DRONE_RPC_SKIP_VERIFY=true
- DRONE_RUNNER_NAME=node12
- DRONE_RUNNER_PRIVILEGED_IMAGES=docker,plugins/docker
privileged: true
volumes:
- /var/run/docker.sock:/var/run/docker.sock

 ```

 `DRONE_RPC_HOST` : drone 访问地址

