package notice

import "resty.dev/v3"

type WechatService struct {
	url string
}

func NewWechatService(url string) *WechatService {
	return &WechatService{
		url: url,
	}
}

func (w *WechatService) Notify(content string) error {
	_, err := resty.New().R().SetBody(map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content": content,
		},
	}).Post(w.url)
	return err
}
