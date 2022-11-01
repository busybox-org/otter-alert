package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xmapst/otter-alert/engine"
	"github.com/xmapst/otter-alert/internal/config"
	"github.com/xmapst/otter-alert/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var rootCmd = &cobra.Command{
	Use:               os.Args[0],
	Version:           VersionIfo(),
	Short:             "A simple otter monitor alert",
	DisableAutoGenTag: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Title, VersionIfo())
		e := engine.New()
		e.Run()
	},
}

func init() {
	time.Local = time.FixedZone("CST", 3600*8)
	registerSignalHandlers()
	logrus.SetFormatter(&utils.ConsoleFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	rootCmd.PersistentFlags().DurationVarP(&config.App.Interval, "interval", "", 5*time.Minute, "monitoring interval. optional")
	rootCmd.PersistentFlags().StringSliceVarP(&config.App.Zookeeper, "zookeeper", "", nil, "connection zookeeper address string")
	_ = rootCmd.MarkPersistentFlagRequired("zookeeper")
	rootCmd.PersistentFlags().StringVarP(&config.App.Notification.Type, "notification.type", "", "dingtalk", "notification send type")
	_ = rootCmd.MarkPersistentFlagRequired("notification.type")
	rootCmd.PersistentFlags().StringVarP(&config.App.Notification.Url, "notification.url", "", "https://oapi.dingtalk.com/robot/send?access_token=xxxxx-xxxxxxxx-xxxxxxxxxx", "notification send url")
	_ = rootCmd.MarkPersistentFlagRequired("notification.url")
	rootCmd.PersistentFlags().StringVarP(&config.App.Notification.Secret, "notification.secret", "", "SECxxxxxxxxxxxxxxxxxxxxx", "notification send secret")
	rootCmd.PersistentFlags().StringVarP(&config.App.Manager.DatabaseUrl, "manager.database", "", "root:123456@tcp(localhost:3306)/otter?charset=utf8&parseTime=True&loc=Local",
		"otter manager database address string")
	rootCmd.PersistentFlags().StringVarP(&config.App.Manager.Endpoint, "manager.endpoint", "", "http://127.0.0.1:8080", "otter manager endpoint")
	_ = rootCmd.MarkPersistentFlagRequired("manager.endpoint")
	rootCmd.PersistentFlags().StringVarP(&config.App.Manager.Username, "manager.username", "", "admin", "otter manager login username")
	rootCmd.PersistentFlags().StringVarP(&config.App.Manager.Password, "manager.password", "", "admin", "otter manager login password")
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
