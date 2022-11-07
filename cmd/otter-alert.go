package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xmapst/otteralert"
	"github.com/xmapst/otteralert/internal/config"
	"github.com/xmapst/otteralert/internal/engine"
	"github.com/xmapst/otteralert/internal/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const Name = "otter-alert"

var rootCmd = &cobra.Command{
	Use:               Name,
	Version:           otteralert.VersionIfo(),
	Short:             "A simple otter alert monitor",
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(otteralert.Title, otteralert.VersionIfo())
		e := engine.New()
		e.Run()
	},
}

func init() {
	var cstSh, err = time.LoadLocation(os.Getenv("TZ"))
	if err != nil {
		cstSh = time.FixedZone("CST", 8*3600)
	}
	time.Local = cstSh
	registerSignalHandlers()
	logrus.SetFormatter(&utils.ConsoleFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	rootCmd.PersistentFlags().DurationVarP(
		&config.App.Interval,
		"interval",
		"",
		5*time.Minute,
		"monitoring interval. \n--interval=5m",
	)

	rootCmd.PersistentFlags().StringSliceVarP(
		&config.App.Zookeeper,
		"zookeeper",
		"",
		nil,
		"connection zookeeper address string. \n--zookeeper=zk-node-1:2181,zk-node-2:2181,zk-node-3:2181",
	)
	_ = rootCmd.MarkPersistentFlagRequired("zookeeper")

	rootCmd.PersistentFlags().StringVarP(
		&config.App.Notification.Type,
		"notification.type",
		"",
		"",
		"notification send type. \n--notification.type=dingtalk",
	)
	_ = rootCmd.MarkPersistentFlagRequired("notification.type")

	rootCmd.PersistentFlags().StringVarP(
		&config.App.Notification.Url,
		"notification.url",
		"",
		"",
		"notification send url. \n--notification.url=https://oapi.dingtalk.com/robot/send?access_token=xxxxx",
	)

	rootCmd.PersistentFlags().StringVarP(
		&config.App.Notification.Secret,
		"notification.secret",
		"",
		"",
		"notification send secret. \n--notification.secret=SEC-xxx",
	)

	rootCmd.PersistentFlags().StringVarP(
		&config.App.Manager.DatabaseUrl,
		"manager.database",
		"",
		"",
		"otter manager database address string. \n--manager.database=user:pass@tcp(127.0.0.1:3306)/otter?charset=utf8&parseTime=True&loc=Local",
	)
	_ = rootCmd.MarkPersistentFlagRequired("manager.database")

	rootCmd.PersistentFlags().StringVarP(
		&config.App.Manager.Endpoint,
		"manager.endpoint",
		"",
		"",
		"otter manager endpoint. \n--manager.endpoint=http://127.0.0.1:8080")
	_ = rootCmd.MarkPersistentFlagRequired("manager.endpoint")

	rootCmd.PersistentFlags().StringVarP(
		&config.App.Manager.Username,
		"manager.username",
		"",
		"",
		"otter manager login username. \n--manager.username=admin",
	)
	_ = rootCmd.MarkPersistentFlagRequired("manager.username")

	rootCmd.PersistentFlags().StringVarP(
		&config.App.Manager.Password,
		"manager.password",
		"",
		"",
		"otter manager login password. \n--manager.password=admin",
	)
	_ = rootCmd.MarkPersistentFlagRequired("manager.password")
}

func registerSignalHandlers() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigs
		os.Exit(0)
	}()
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}
