package svc

import (
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
