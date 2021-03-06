package env

import (
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// AppName holds the name of this application
	AppName = "CPMA"

	// logFile - keeps full path to the logging file
	logFile = "cpma.log.json" // TODO: we may want this configurable
)

var (
	// ConfigFile - keeps full path to the configuration file
	ConfigFile string

	viperConfig *viper.Viper
)

func init() {
	viperConfig = viper.New()
}

// Config returns pointer to the viper config
func Config() *viper.Viper {
	return viperConfig
}

// InitConfig initializes application's configuration
func InitConfig() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		logrus.WithError(err).Fatal("Can't detect home user directory")
	}
	viperConfig.Set("home", home)

	if ConfigFile != "" {
		// Use config file from the flag.
		viperConfig.SetConfigFile(ConfigFile)
	} else {
		// TODO: Do we want default config file at all?
		// Search config in home directory with name ".cpma" (without extension).
		viperConfig.AddConfigPath(home)
		viperConfig.SetConfigName(".cpma")
	}

	// Fill in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viperConfig.ReadInConfig(); err != nil {
		logrus.Fatal(err)
	}
}

// InitLogger initializes stderr and logger to file
func InitLogger() {
	logLevel := logrus.InfoLevel
	if viperConfig.GetBool("debug") {
		logLevel = logrus.DebugLevel
		logrus.SetReportCaller(true)
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})
	logrus.SetLevel(logLevel)
	logrus.Debugf("%s is running in debug mode", AppName)

	fileHook, _ := NewLogFileHook(
		LogFileConfig{
			Filename: logFile,
			MaxSize:  5, // MiB
			Level:    logLevel,
			Formatter: &logrus.JSONFormatter{
				PrettyPrint: true,
			},
		},
	)
	logrus.AddHook(fileHook)
}
