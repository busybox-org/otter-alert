# otter-alert

阿里巴巴分布式数据库同步系统监控告警

[x] 状态告警  
[x] 日志告警  
[x] 延时自愈  
[x] 自动解挂

## Deployment
```shell
kubectl apply -f https://raw.githubusercontent.com/busybox-org/otter-alert/main/k8s.yaml
```

## HELP
```shell
# ./otter-alert --help
A simple otter alert monitor

Usage:
  ./bin/otter-alert [flags]

Flags:
      --alert_ak string            Access key for alerting (required)
      --alert_sk string            Secret key for alerting (Optional)
      --cron string                Cron expression for automatic execution (Optional) (default "*/30 * * * * *")
  -h, --help                       help for ./bin/otter-alert
      --manager.database string    Otter manager database address string (default "username:password@tcp(host:port)/database?charset=utf8&parseTime=True&loc=Local")
      --manager.endpoint string    Otter manager web endpoint (default "http://username:password@host:port")
      --manager.zookeeper string   Otter manager zookeeper address (default "host:port,host:port,host:port")
```