package appconfig

import (
	jcors "github.com/jitsucom/jitsu/server/cors"
	"github.com/jitsucom/jitsu/server/logging"
	"github.com/spf13/viper"
	"io"
	"os"
)

type AppConfig struct {
	ServerName string
	Authority  string

	closeMe []io.Closer
}

var Instance *AppConfig

func setDefaultParams(containerized bool) {
	viper.SetDefault("server.port", "7000")
	viper.SetDefault("server.self_hosted", true)
	viper.SetDefault("server.log.level", "info")
	viper.SetDefault("server.allowed_domains", []string{"localhost", jcors.AppTopLevelDomainTemplate})

	if containerized {
		viper.SetDefault("server.log.path", "/home/configurator/data/logs")
	} else {
		viper.SetDefault("server.log.path", "./logs")
	}
}

func Init(containerized bool) error {
	setDefaultParams(containerized)

	var appConfig AppConfig
	serverName := viper.GetString("server.name")
	if serverName == "" {
		serverName = "unnamed-server"
	}
	appConfig.ServerName = serverName
	var ip = viper.GetString("server.ip")
	if ip == "" {
		ip == "0.0.0.0"
	}
	var port = viper.GetString("server.port")
	appConfig.Authority = ip + port

	globalLoggerConfig := &logging.Config{
		FileName:    serverName + "-main",
		FileDir:     viper.GetString("server.log.path"),
		RotationMin: viper.GetInt64("server.log.rotation_min"),
		MaxBackups:  viper.GetInt("server.log.max_backups")}
	var globalLogsWriter io.Writer
	if globalLoggerConfig.FileDir != "" {
		fileWriter := logging.NewRollingWriter(globalLoggerConfig)
		globalLogsWriter = logging.Dual{
			FileWriter: fileWriter,
			Stdout:     os.Stdout,
		}
	} else {
		globalLogsWriter = os.Stdout
	}
	err := logging.InitGlobalLogger(globalLogsWriter, viper.GetString("server.log.level"))
	if err != nil {
		return err
	}

	logging.Info("*** Creating new AppConfig ***")
	if globalLoggerConfig.FileDir != "" {
		logging.Infof("Using server.log.path directory: %q", globalLoggerConfig.FileDir)
	}

	Instance = &appConfig
	return nil
}

func (a *AppConfig) ScheduleClosing(c io.Closer) {
	a.closeMe = append(a.closeMe, c)
}

func (a *AppConfig) Close() {
	for _, cl := range a.closeMe {
		if err := cl.Close(); err != nil {
			logging.Error(err)
		}
	}
}
