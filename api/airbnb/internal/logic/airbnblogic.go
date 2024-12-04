package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/svc"
	"github.com/Cosmosmay/ctw_zy/api/airbnb/internal/types"
	tw_model "github.com/Cosmosmay/ctw_zy/model/mysql"

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

func (l *AirbnbLogic) Producer(ctx context.Context, rdbCli *redis.Client, queueName string, dataFile string) error {
	file, err := os.ReadFile(dataFile)
	if err != nil {
		log.Fatalf("Error reading tasks file: %v\n", err)
	}

	var tasks types.TaskList
	if err := json.Unmarshal(file, &tasks); err != nil {
		log.Fatalf("Error parsing tasks: %v\n", err)
	}

	for _, task := range tasks.Tasks {
		taskData, err := json.Marshal(task)
		if err != nil {
			log.Fatalf("Error serializing task: %v\n", err)
			continue
		}

		fmt.Printf("Producing task: %v\n", task.Name)

		// 使用 LPUSH 将任务添加到 Redis 队列
		err = rdbCli.LPush(ctx, queueName, taskData).Err()
		if err != nil {
			log.Fatalf("Error pushing task to Redis queue: %v\n", err)
			continue
		}
	}
	return nil
}

func (l *AirbnbLogic) Consumer(ctx context.Context, rdbCli *redis.Client, queueName string, workers int, dbCli sqlx.SqlConn) error {
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				// 使用 RPOP 从 Redis 队列中取出任务
				fmt.Println(queueName)
				taskData, err := rdbCli.RPop(ctx, queueName).Bytes()
				if err != nil {
					if err == redis.Nil {
						fmt.Println("Queue is empty, waiting for tasks...")
					} else {
						log.Printf("Error popping task from Redis: %v\n", err)
					}
					time.Sleep(1 * time.Second) // 如果队列为空，等待并重试
					continue
				}

				fmt.Printf("Worker %d processing task: %s\n", workerID, taskData)
				var task types.Task
				err = json.Unmarshal(taskData, &task)
				if err != nil {
					log.Printf("Error unmarshalling task data: %v\n", err)
					continue
				}

				fmt.Printf("Worker %d processing task: %v\n", workerID, task.Name)
				scrapePage(task, dbCli)
			}
		}(i)
	}
	wg.Wait()
	return nil
}

// 抓去页面数据
func scrapePage(task types.Task, dbCli sqlx.SqlConn) error {
	twModel := tw_model.NewAirbnbInfoModel(dbCli)
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true), //设置成无浏览器弹出模式
		chromedp.Flag("blink-settings", "imageEnable=false"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"),
	}
	c, _ := chromedp.NewExecAllocator(context.Background(), options...)
	chromeCtx, cancel := chromedp.NewContext(c, chromedp.WithLogf(log.Printf))
	_ = chromedp.Run(chromeCtx, make([]chromedp.Action, 0, 1)...)
	timeOutCtx, cancel := context.WithTimeout(chromeCtx, 60*time.Second)
	defer cancel()
	var htmlContent string
	err := chromedp.Run(timeOutCtx,
		chromedp.Navigate(task.URL),
		//需要爬取的网页的url
		//chromedp.WaitVisible(`#content > div > section.fp-tournament-award-badge-carousel_awardBadgeCarouselSection__w_Ys5 > div > div > div.col-12.fp-tournament-award-badge-carousel_awardCarouselColumn__fQJLf.g-0 > div > div > div > div > div > div > div.slick-slide.slick-active.slick-current > div > div > div`),
		//chromedp.WaitVisible(`[data-testid="change-dates-checkIn"]`),
		//等待某个特定的元素出现
		chromedp.OuterHTML(`document.querySelector("body")`, &htmlContent, chromedp.ByJSPath),
		//生成最终的html文件并保存在htmlContent文件中
	)
	if err != nil {
		log.Fatal(err)
		return err
	}
	//log.Println(htmlContent)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Fatal("Failed to parse HTML: ", err)
		return err
	}

	doc.Find("div[data-testid='card-container']").Each(func(i int, s *goquery.Selection) {
		s.Find("a[href]").Each(func(i int, s *goquery.Selection) {
			// 获取 href 属性值
			href, exists := s.Attr("href")
			if !exists {
				return
			}
			task.URL = "https://zh.airbnb.com/" + href
			scrapePage(task, dbCli)
		})
	})

	insertData := new(tw_model.AirbnbInfo)
	insertData.AirbnbUrl = task.URL
	//hotelName
	hotelName := doc.Find(`[elementtiming="LCP-target"]`).Text()
	if hotelName != "" {
		insertData.HotelName = hotelName
	}
	//price
	price := doc.Find("div._1jo4hgw > span._11jcbg2").Text()
	if price != "" {
		price = strings.Replace(price, "¥", "", -1) // 去除人民币符号
		price = strings.Replace(price, ",", "", -1) // 去除千位分隔符
		// 将剩余的字符串转换为整数
		priceInt, err := strconv.Atoi(price)
		if err != nil {
			log.Fatalf("Error converting price to int: %v\n", err)
			return err
		}
		insertData.Price = float64(priceInt)
	}

	//checkIn and checkOut
	checkIn := doc.Find(`[data-testid="change-dates-checkIn"]`).Text()
	checkOut := doc.Find(`[data-testid="change-dates-checkOut"]`).Text()
	if checkIn != "" && checkOut != "" {
		// 解析日期
		layout := "2006/01/02" // 定义日期格式
		checkIndate, _ := time.Parse(layout, checkIn)
		checkOutDate, _ := time.Parse(layout, checkOut)
		insertData.CheckInDate = checkIndate
		insertData.CheckOutDate = checkOutDate
	}

	// 房客
	guests := doc.Find(`div.__1e5z145 > span._j1kt73`).Text()
	if guests != "" {
		re := regexp.MustCompile(`\d+`)
		numStr := re.FindString(guests)
		// 将提取的数字字符串转换为整数
		var number int
		_, err := fmt.Sscanf(numStr, "%d", &number)
		if err != nil {
			log.Fatal("convert failed:", err)
			return err
		}
		insertData.Guests = int64(number)
	}

	//star
	star := doc.Find(`[data-testid="pdp-reviews-highlight-banner-host-rating"] > span`).Text()
	if star != "" {
		starParse, err := strconv.Atoi(star)
		if err != nil {
			log.Fatalf("Error converting price to int: %v\n", err)
			return err
		}
		insertData.Star = int64(starParse)
	}

	twModel.Insert(*insertData)
	return nil
}
