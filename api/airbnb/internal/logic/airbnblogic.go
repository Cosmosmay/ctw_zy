package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/svc"
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AirbnbLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAirbnbLogic(ctx context.Context, svcCtx *svc.ServiceContext) AirbnbLogic {
	return AirbnbLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AirbnbLogic) Producer(taskName string, queue *TaskQueue) error {
	file, err := os.Open(taskName)
	if err != nil {
		log.Fatalf("Error opening task file: %v", err)
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("Error reading task file: %v", err)
		return err
	}

	var tasks types.TaskList
	if err := json.Unmarshal(data, &tasks); err != nil {
		log.Fatalf("Error parsing task file: %v", err)
	}

	for _, task := range tasks.Tasks {
		queue.Enqueue(task)
	}
	queue.Close()

	return nil
}

func (l *AirbnbLogic) Consumer(workers int, queue *TaskQueue) error {
	//todo
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for task := range queue.queue {
				fmt.Printf("Worker %d processing task: %v\n", workerID, task.URL)
				scrapePage(task)
			}
		}(i)
	}
	wg.Wait()
	return nil
}

func scrapePage(task types.Task) error {

	return nil
}
