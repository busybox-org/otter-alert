package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
	"github.com/xmapst/logx"

	"github.com/busybox-org/otteralert/internal/core"
)

func main() {
	root := &cobra.Command{
		Use:           os.Args[0],
		Short:         "A simple otter alert monitor",
		Long:          "A simple otter alert monitor",
		SilenceUsage:  true,
		SilenceErrors: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var cstSh, err = time.LoadLocation(os.Getenv("TZ"))
			if err != nil {
				cstSh = time.FixedZone("CST", 8*3600)
			}
			time.Local = cstSh
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := filepath.Abs(os.Args[0])
			if err != nil {
				logx.Errorln(err)
				return err
			}
			svc, err := service.New(core.New(cmd.Flags()), &service.Config{
				Name:        name,
				DisplayName: name,
				Description: "Operating System Remote Executor Api",
				Executable:  name,
				Arguments:   os.Args[1:],
			})
			if err != nil {
				return err
			}
			err = svc.Run()
			if err != nil {
				return err
			}
			return nil
		},
	}

	// alert flags
	root.Flags().String("alert_ak", "", "Access key for alerting (required)")
	_ = root.MarkFlagRequired("alert_ak")
	root.Flags().String("alert_sk", "", "Secret key for alerting (Optional)")

	// cron flags
	root.Flags().String("cron", "*/30 * * * * *", "Cron expression for automatic execution (Optional)")

	// manager flags
	root.Flags().String("manager.database", "username:password@tcp(host:port)/database?charset=utf8&parseTime=True&loc=Local", "Otter manager database address string")
	root.Flags().String("manager.endpoint", "http://username:password@host:port", "Otter manager web endpoint")
	root.Flags().String("manager.zookeeper", "host:port,host:port,host:port", "Otter manager zookeeper address")
	if err := root.Execute(); err != nil {
		logx.Fatalln(err)
	}
}
