# otter-alert

阿里巴巴分布式数据库同步系统监控告警

-[x] 发送异常信息到钉钉群机器人  
-[x] 延迟过大报警以及恢复正常通知  
-[x] 异常挂起抓取日志信息以及自动解挂  
-[x] binlog位置点显示  
-[x] 源库信息显示  
-[x] 同步与恢复时间显示  

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