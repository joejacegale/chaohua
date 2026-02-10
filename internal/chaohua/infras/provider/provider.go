package provider

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/joejacegale/chaohua/internal/chaohua/domain"

	"resty.dev/v3"
)

type ChaohuaProvider struct {
	client *resty.Client
}

const userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"

func new(sub string) *ChaohuaProvider {
	return &ChaohuaProvider{
		client: resty.New().
			SetHeader("User-Agent", userAgent).
			SetCookie(&http.Cookie{Name: "SUB", Value: sub}).
			SetHeader("Referer", "https://weibo.com"),
	}
}

func NewChaohuaProvider(keyStr, file string) (*ChaohuaProvider, error) {
	cs, err := NewCookieService(keyStr, file)
	if err != nil {
		return nil, err
	}

	sub, err := cs.GetSub()
	if err != nil {
		return nil, err
	}

	return new(sub), nil
}

func (p *ChaohuaProvider) R() *resty.Request {
	return p.client.R()
}

func (p *ChaohuaProvider) ChaoHuaList() ([]*domain.ChaoHua, error) {
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
	req := p.R().SetQueryParam("tabid", "231093_-_chaohua")
	var ids []*domain.ChaoHua
	maxPage := 1
	for page := 1; page <= maxPage; page++ {
		var res response
		_, err := req.SetQueryParam("page", strconv.Itoa(page)).SetResult(&res).Get("https://weibo.com/ajax/profile/topicContent")
		if err != nil || res.Ok == 0 {
			return nil, errors.New("超话列表获取失败，登录状态过期啦")
		}
		maxPage = res.Data.Max_page
		for _, item := range res.Data.List {
			arr := strings.Split(item.Link, "/")
			ids = append(ids, domain.NewChaoHua(arr[len(arr)-1], item.Title))
		}
	}
	return ids, nil
}

func (p *ChaohuaProvider) Sign(ch *domain.ChaoHua) error {
	var res struct {
		Code any
		Msg  string
	}
	_, err := p.R().
		SetQueryParam("api", "http://i.huati.weibo.com/aj/super/checkin").
		SetQueryParam("id", ch.ID()).
		SetResult(&res).
		Get("https://weibo.com/p/aj/general/button")
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
		return errors.New(`"` + ch.Title() + `"` + res.Msg)
	}
	return nil
}
