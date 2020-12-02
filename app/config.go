package app

import (
	"github.com/spf13/viper"
)

type Config struct {
	*viper.Viper
}
