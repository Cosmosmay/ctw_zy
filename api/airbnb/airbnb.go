package main

import (
	"context"
	"flag"
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/logic"
	"log"

	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/config"
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
)

var (
	mode       = flag.String("mode", "consumer/producer", "run mode")
	workers    = flag.Int("workers", 10, "number of workers")
	data       = flag.String("data", "tasks.json", "data for task")
	queue      = flag.String("queue", "airbnb", "queue name")
	configFile = flag.String("f", "etc/airbnb-api.yaml", "the config file")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	ctx := svc.NewServiceContext(c)
	db, err := config.SetupDatabase(c.Database)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer db.Close()

	l := logic.NewAirbnbLogic(context.Background(), ctx)

	if *mode == "producer" {
		//todo
		l.Producer()
	} else if *mode == "consumer" {
		//todo
		l.Consumer()
	}
}
