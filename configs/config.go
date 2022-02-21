package configs

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func InitConfig() error {
	//viper init
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	//env init
	if err := godotenv.Load(); err != nil {
		return err
	}

	return nil
}
