package init

import (
	ErrH "errorhandle"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
)

func Config() (err error) {
	var (
		pwd string
	)

	viper.New()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")

	//watch the config change
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf(ErrH.ErrorLog(6140), fmt.Sprintf("%v", e.Name))
	})

	if pwd, err = os.Getwd(); err != nil {
		os.Exit(1)
		return
	}
	log.Printf(ErrH.ErrorLog(6141), fmt.Sprintf("%v", pwd))

	//Find and read the config and token file
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	return

}
