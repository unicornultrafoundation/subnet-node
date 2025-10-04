package common

import (
	"time"

	"github.com/spf13/viper"
)

type OperatorConfig struct {
	PruneInterval      time.Duration
	WebRefreshInterval time.Duration
	RetryDelay         time.Duration
	ProviderAddress    string
}

func GetOperatorConfigFromViper() OperatorConfig {
	return OperatorConfig{
		PruneInterval:      viper.GetDuration(FlagPruneInterval),
		WebRefreshInterval: viper.GetDuration(FlagWebRefreshInterval),
		RetryDelay:         viper.GetDuration(FlagRetryDelay),
		ProviderAddress:    viper.GetString(flagProviderAddress),
	}
}
