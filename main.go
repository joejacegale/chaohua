package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"resty.dev/v3"
)

func main() {
	notify(do(sub))
}

var sub string
var webhook string
var mobile string

func init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		log.Fatal(err)
	}
	file, err := root.OpenFile("config.yml", os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	dec := yaml.NewDecoder(file)
	var config struct {
		Sub     string
		Webhook string
		Mobile  string
	}
	if err := dec.Decode(&config); err != nil {
		log.Fatal(err)
	}
	sub = config.Sub
	webhook = config.Webhook
	mobile = config.Mobile
}

func notify(err error) {
	resty.New().R().SetBody(map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content":               err.Error(),
			"mentioned_mobile_list": []string{mobile},
		},
	}).Post(webhook)
}

func do(sub string) error {
	chaohuas, err := getChaoHuaList(sub)
	if err != nil {
		return err
	}
	total, success, fail := len(chaohuas), 0, 0
	for _, ch := range chaohuas {
		err := sign(ch, sub)
		if err != nil {
			fail++
		} else {
			success++
		}
		time.Sleep(time.Duration(rand.Intn(500) + 500))
	}
	return fmt.Errorf("一共 %d 个超话\n签到成功 %d 个\n签到失败 %d 个", total, success, fail)
}

type chaohua struct {
	id    string
	title string
}

func getChaoHuaList(sub string) (ids []*chaohua, err error) {
	type response struct {
		Ok   int
		Data struct {
			Total_number int
			Max_page     int
			List         []struct {
				Title string
				Link  string
			}
		}
	}
	client := resty.New()
	client.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36").
		SetHeader("Referer", "https://weibo.com").
		SetCookie(&http.Cookie{Name: "SUB", Value: sub}).
		SetQueryParam("tabid", "231093_-_chaohua")

	maxPage := 1
	for page := 1; page <= maxPage; page++ {
		var res response
		_, err := client.R().SetQueryParam("page", strconv.Itoa(page)).SetResult(&res).Get("https://weibo.com/ajax/profile/topicContent")
		if err != nil || res.Ok == 0 {
			return nil, errors.New("超话列表获取失败，登录状态过期啦")
		}
		maxPage = res.Data.Max_page
		for _, item := range res.Data.List {
			arr := strings.Split(item.Link, "/")
			ids = append(ids, &chaohua{arr[len(arr)-1], item.Title})
		}
	}
	return
}

func sign(ch *chaohua, sub string) error {
	type response struct {
		Code any
		Msg  string
	}
	client := resty.New()
	client.SetHeader("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36").
		SetCookie(&http.Cookie{Name: "SUB", Value: sub})
	var res response
	_, err := client.R().SetQueryParam("api", "http://i.huati.weibo.com/aj/super/checkin").SetQueryParam("id", ch.id).SetResult(&res).Get("https://weibo.com/p/aj/general/button")
	if err != nil {
		return err
	}
	var code int
	switch res.Code.(type) {
	case int:
		code = res.Code.(int)
	case string:
		code, _ = strconv.Atoi(res.Code.(string))
	}
	if code != 100000 {
		return errors.New(`"` + ch.title + `"` + res.Msg)
	}
	return nil
}
