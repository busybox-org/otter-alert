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
# ./otteralert --help
A simple otter monitor alert

Usage:
  ./otteralert [flags]

Flags:
  -h, --help                         help for ./otteralert
      --interval duration            monitoring interval. optional (default 5m0s)
      --manager.database string      otter manager database address string (default "root:123456@tcp(localhost:3306)/otter?charset=utf8&parseTime=True&loc=Local")
      --manager.endpoint string      otter manager endpoint (default "http://127.0.0.1:8080")
      --manager.password string      otter manager login password (default "admin")
      --manager.username string      otter manager login username (default "admin")
      --notification.secret string   notification send secret (default "SECxxxxxxxxxxxxxxxxxxxxx")
      --notification.type string     notification send type (default "dingtalk")
      --notification.url string      notification send url (default "https://oapi.dingtalk.com/robot/send?access_token=xxxxx-xxxxxxxx-xxxxxxxxxx")
  -v, --version                      version for ./otteralert
      --zookeeper strings            connection zookeeper address string
```