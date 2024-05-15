package config

import "os"

type Config struct {
	Address                   string
	FollowerServiceAddress    string
	StakeholderServiceAddress string
	GreeterServiceAddress     string
}

func GetConfig() Config {
	return Config{
		GreeterServiceAddress:     os.Getenv("GREETER_SERVICE_ADDRESS"),
		FollowerServiceAddress:    os.Getenv("FOLLOWER_SERVICE_ADDRESS"),
		StakeholderServiceAddress: os.Getenv("STAKEHOLDER_SERVICE_ADDRESS"),
		Address:                   os.Getenv("GATEWAY_ADDRESS"),
	}
}
