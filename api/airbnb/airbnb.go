package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/config"
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/logic"
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/svc"
	"log"
	"os"
)

var (
	mode       = flag.String("mode", "", "run mode: consumer|producer")
	workers    = flag.Int("workers", 10, "number of workers (for consumer mode)")
	taskFile   = flag.String("data", "etc/tasks.json", "task file for producer")
	queue      = flag.String("queue", "", "queue name")
	configFile = flag.String("f", "etc/airbnb-api.yaml", "config file path")
)

func main() {
	// 检查命令结构
	if len(os.Args) < 3 || os.Args[1] != "start" {
		fmt.Println("usage: airbnb-cli start [consumer|producer] [--workers=N] [--data=FILE]")
		os.Exit(1)
	}
	// 解析标志参数（跳过前两个参数：命令本身和 start 子命令）
	flag.CommandLine.Parse(os.Args[2:])
	// Load configuration
	c, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	// Initialize context and service context
	ctx := context.Background()
	svcCtx := svc.NewServiceContext(c)

	// Setup database and Redis clients
	dbCli, err := config.SetupDatabase(c.Database)
	if err != nil {
		log.Fatalf("error setting up database: %v", err)
	}
	//defer func() {
	//	if err := dbCli.Close(); err != nil {
	//		log.Printf("warning: error closing database: %v", err)
	//	}
	//}()

	rdbCli, err := config.SetupRedis(c.RedisConfig)
	if err != nil {
		log.Fatalf("error setting up Redis: %v", err)
	}
	defer func() {
		if err := rdbCli.Close(); err != nil {
			log.Printf("warning: error closing Redis: %v", err)
		}
	}()

	// Create logic instance
	l := logic.NewAirbnbLogic(ctx, svcCtx)

	// Run in specified mode
	switch *mode {
	case "producer":
		log.Printf("starting producer with task file: %s", *taskFile)
		if err := l.Producer(ctx, rdbCli, *queue, *taskFile); err != nil {
			log.Fatalf("error in producer: %v", err)
		}
	case "consumer":
		log.Printf("starting consumer with %d workers", *workers)
		if err := l.Consumer(ctx, rdbCli, *queue, *workers, *dbCli); err != nil {
			log.Fatalf("error in consumer: %v", err)
		}
	default:
		log.Fatalf("invalid mode: %s (must be 'consumer' or 'producer')", *mode)
	}
}
