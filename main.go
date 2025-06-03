package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"os"
	"spider/yypp.me/config"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Calculator 类型定义了一个函数，接受一个 int 和一个 *sync.WaitGroup，返回一个 int
type Calculator func(int, *sync.WaitGroup) error

func main() {
	cfg, err := config.LoadConfig("yypp.me/config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	fmt.Println("请输入1 保存html 输入2 html 转json保存 输入3 json 转img保存")
	var choice int

	var operation Calculator

	fmt.Scanln(&choice)
	if choice == 1 {
		// 赋值为 processHtmlPage 函数
		operation = processHtmlPage
		fmt.Println("保存html")
	} else if choice == 2 {
		// 这里应该赋值为 htmlToJson 函数
		operation = htmlToJson
		fmt.Println("html 转json保存")
	} else if choice == 3 {
		// 这里应该赋值为 jsonToImg 函数
		operation = jsonToImg
		fmt.Println("json 转img保存")
	} else {
		fmt.Println("输入错误")
		return
	}

	// 记录开始时间
	startTime := time.Now()
	// 打印开始时间
	fmt.Println("开始时间:", startTime)
	fmt.Printf("开始处理 %s 网站\n", cfg.App.BaseDomainUrl)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// 为每一页启动一个 goroutine
	for page := cfg.App.StartPage; page <= cfg.App.TotalPages; page++ {
		wg.Add(1)

		go func(p int) {

			err := operation(p, &wg)
			if err != nil {
				mu.Lock()
				errors = append(errors, err)
				mu.Unlock()
			}
		}(page)
	}
	// 等待所有 goroutine 完成
	wg.Wait()

	// 处理错误
	if len(errors) > 0 {
		fmt.Println("处理过程中出现以下错误:")
		for _, err := range errors {
			fmt.Println(err)
		}
	}
	fmt.Println("所有页面处理完成")
	endTime := time.Now()
	// 打印结束时间
	fmt.Println("结束时间:", endTime)
	log.Printf("总共耗时 %s", endTime.Sub(startTime))
}

// 处理单个页面的函数
func processHtmlPage(page int, wg *sync.WaitGroup) error {
	defer wg.Done()
	cfg, err := config.LoadConfig("yypp.me/config/config.yaml")
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}

	filePath := fmt.Sprintf("%s%d.html", cfg.App.BaseCachePageURL, page)
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("网页 %d 的内容已存\n", page)
	}

	// 构建当前页的URL
	url := cfg.App.BaseDomainUrl + "list/fuli-" + strconv.Itoa(page) + ".html"
	log.Printf("请求 %s 出错: %v\n", url, err)
	// 发送HTTP GET请求
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("请求 %s 出错: %v\n", url, err)
		// 打印日志文件到 /data/yypp.me.log 文件中
		logfilePath := fmt.Sprintf("%s%s.log", cfg.App.BaseCacheLogURL, "yypp.me")
		f, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("打开日志文件 %s 出错: %v\n", logfilePath, err)
		}
		defer f.Close()
		_, err = f.WriteString(fmt.Sprintf("请求 %s 出错: %v\n", url, err))
		if err != nil {
			log.Printf("写入日志文件 %s 出错: %v\n", logfilePath, err)
		}

		return err
	}
	defer resp.Body.Close()

	// 使用goquery解析HTML文档
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("请求 %s 时状态码错误: %d %s\n", url, resp.StatusCode, resp.Status)
		log.Printf("请求 %s 时状态码错误: %d %s\n", url, resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("解析 %s 的HTML文档出错: %v\n", url, err)
		return err
	}

	// 获取 HTML 代码
	html, err := doc.Html()
	if err != nil {
		fmt.Printf("获取 %s 的HTML代码出错: %v\n", url, err)
		return err
	}

	// 创建 baseCachePageURL [cache/page] 目录（如果不存在）
	err = os.MkdirAll(cfg.App.BaseCachePageURL, os.ModePerm)
	if err != nil {
		fmt.Printf("创建目录 %s 出错: %v\n", cfg.App.BaseCachePageURL, err)
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("创建文件 %s 出错: %v\n", filePath, err)
		return err
	}
	defer f.Close()
	_, err = f.WriteString(html)
	if err != nil {
		fmt.Printf("写入文件 %s 出错: %v\n", filePath, err)
		return err
	}

	log.Printf("Finished processing page %d", page)
	return nil
}

// 示例函数：将 HTML 转换为 JSON
func htmlToJson(page int, wg *sync.WaitGroup) error {
	defer wg.Done()
	cfg, err := config.LoadConfig("yypp.me/config/config.yaml")
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}
	// 判断文件是否存在 cache/page/{i}.html 文件中
	// list 转成 json 存入文件 ./cache/data/1.json
	localJsonPath := cfg.App.BaseCacheDataURL + strconv.Itoa(page) + ".json"
	filePath := fmt.Sprintf(localJsonPath)
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("done网页 %s 的内容已存\n", filePath, err)
	}

	localPath := cfg.App.BaseCachePageURL + strconv.Itoa(page) + ".html"

	// 获取 localPath 文件内容
	content, err := os.ReadFile(localPath)
	if err != nil {
		log.Printf("读取文件 %s 失败: %w", localPath, err)
		return nil
	}

	// 解析 HTML 内容
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(content)))
	if err != nil {
		log.Printf("解析 HTML 文件 %s 失败: %w", localPath, err)
		return nil
	}

	//fmt.Println(doc.Html())// 打印 doc
	// 解析列表标题，id，链接，内容,图片链接
	list := parseList(doc)

	fmt.Println("正在处理第", page, "页")
	err = saveListToJSON(list, localJsonPath)
	if err != nil {
		log.Printf("保存 JSON 文件 %s 失败: %w", localJsonPath, err)
	}
	return nil
}

// 示例函数：将 JSON 转换为图片
func jsonToImg(page int, wg *sync.WaitGroup) error {
	defer wg.Done()
	cfg, err := config.LoadConfig("yypp.me/config/config.yaml")
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}

	jsonFilePath := cfg.App.BaseCacheDataURL + strconv.Itoa(page) + ".json"

	// 读取 JSON 文件内容
	data, err := readJSONFile(jsonFilePath)
	if err != nil {
		return err
	}

	// 创建保存图片的目录
	if err := createImageDirectory(cfg.App.BaseWaterPageUrl); err != nil {
		return err
	}

	// 遍历数组，获取图片地址并下载图片
	for i, item := range data {
		imagename := fmt.Sprintf("img_%d-%d-%d.jpg", page, i+1, i+1)
		// 判断图片是否已经下载过
		if _, err := os.Stat(fmt.Sprintf("%s/%s", cfg.App.BaseWaterPageUrl, imagename)); !os.IsNotExist(err) {
			log.Printf("Image %d from page %d already exists, skipping\n", i+1, page)
			continue
		}
		// 判断item["img"] 是否有是完整的URL，如果没有，则加上域名
		// 因为JSON文件中图片地址可能是相对路径，所以需要加上域名
		if !strings.HasPrefix(item["img"], "http") {
			item["img"] = cfg.App.BaseDomainUrl + item["img"]
		}

		imgURL := item["img"]
		if imgURL != "" {
			// 生成保存图片的文件名

			savePath := fmt.Sprintf("%s/%s", cfg.App.BaseWaterPageUrl, imagename)
			log.Printf("Downloading image %d from %s to %s\n", i+1, imgURL, savePath)
			err := downloadImage(imgURL, savePath)
			if err != nil {

				log.Printf("Failed to download image %d: %v\n", i+1, err)
				// 记录错误日志
				logFile, err := os.OpenFile(cfg.App.BaseCacheLogURL+"error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Printf("Failed to open error log file: %v\n", err)
					continue
				}
				defer logFile.Close()
				logFile.WriteString(fmt.Sprintf("Failed to download image %d from %s to %s: %v\n", i+1, imgURL, savePath, err))
				continue
			}
		}
	}

	log.Printf("Finished processing page %d", page)
	return nil
}

// 将列表信息保存为 JSON 文件
func saveListToJSON(list []map[string]string, filePath string) error {
	jsonStr, err := json.Marshal(list)
	if err != nil {
		log.Printf("将列表转换为 JSON 失败: %w", err)
	}

	jsonFile, err := os.Create(filePath)
	if err != nil {
		log.Printf("创建文件 %s 失败: %w", filePath, err)
	}
	defer jsonFile.Close()

	_, err = jsonFile.Write(jsonStr)
	if err != nil {
		log.Printf("写入文件 %s 失败: %w", filePath, err)
	}
	return nil
}

// 解析 HTML 文档中的列表信息
func parseList(doc *goquery.Document) []map[string]string {
	cfg, err := config.LoadConfig("yypp.me/config/config.yaml")
	if err != nil {
		log.Printf("加载配置失败: %v", err)
	}
	var list []map[string]string
	doc.Find(".stui-vodlist__box").Each(func(i int, selection *goquery.Selection) {
		title := selection.Find(".title a").Text()
		link, _ := selection.Find(".title a").Attr("href")
		img, _ := selection.Find(".stui-vodlist__thumb").Attr("data-original")
		tagdesc := selection.Find(".pic-text.text-right").Text()
		var tags []string
		selection.Find(".pic-text1.text-right.hidden-xs").Find("b").Each(func(j int, s *goquery.Selection) {
			bText := s.Text()
			tags = append(tags, strings.TrimSpace(bText))
		})
		item := map[string]string{
			"title":   title,
			"link":    link,
			"img":     img,
			"tags":    strings.Join(tags, ", "),
			"tagdesc": tagdesc,
			"baseUrl": cfg.App.BaseDomainUrl,
		}
		list = append(list, item)
	})
	return list
}

// 从 JSON 文件中读取数据并转换为数组
func readJSONFile(filePath string) ([]map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer file.Close()

	var data []map[string]string
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JSON file: %w", err)
	}
	return data, nil
}

// 下载图片并保存到本地
func downloadImage(url string, savePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image from %s: status code %d", url, resp.StatusCode)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", savePath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy image data to %s: %w", savePath, err)
	}
	return nil
}

// 创建保存图片的目录
func createImageDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create image directory %s: %w", dir, err)
		}
	}
	return nil
}
