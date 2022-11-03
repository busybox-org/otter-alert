# otter-alert

阿里巴巴分布式数据库同步系统监控告警

[x] 状态告警  
[x] 日志告警  
[x] 延时自愈  
[x] 自动解挂

## Deployment
```shell
kubectl apply -f https://raw.githubusercontent.com/xmapst/otter-alert/main/k8s.yaml
```

## HELP
```shell
# ./otter-alert --help
A simple otter alert monitor

Usage:
  otter-alert [flags]

Flags:
  -h, --help                         help for otter-alert
      --interval duration            monitoring interval. 
                                     --interval=5m (default 5m0s)
      --manager.database string      otter manager database address string. 
                                     --manager.database=mysql://root:123456@localhost:3306/otter?charset=utf8&parseTime=True&loc=Local
      --manager.endpoint string      otter manager endpoint. 
                                     --manager.endpoint=http://127.0.0.1:8080
      --manager.password string      otter manager login password. 
                                     --manager.password=admin
      --manager.username string      otter manager login username. 
                                     --manager.username=admin
      --notification.secret string   notification send secret. 
                                     --notification.secret=SEC-xxx
      --notification.type string     notification send type. 
                                     --notification.type=dingtalk
      --notification.url string      notification send url. 
                                     --notification.url=https://oapi.dingtalk.com/robot/send?access_token=xxxxx
  -v, --version                      version for otter-alert
      --zookeeper strings            connection zookeeper address string. 
                                     --zookeeper=zk-node-1:2181,zk-node-2:2181,zk-node-3:2181

```