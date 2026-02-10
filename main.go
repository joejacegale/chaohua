package main

import (
	"os"

	"github.com/joejacegale/chaohua/internal/chaohua/app"
	"github.com/joejacegale/chaohua/internal/chaohua/domain"
	"github.com/joejacegale/chaohua/internal/chaohua/infras/notice"
	"github.com/joejacegale/chaohua/internal/chaohua/infras/provider"
)

func main() {
	keyStr := os.Getenv("COOKIE_SECRET")
	file := os.Getenv("COOKIE_FILE")

	cs, err := provider.NewChaohuaProvider(keyStr, file)
	if err != nil {
		panic(err)
	}

	webhook := os.Getenv("WEBHOOK")

	notifyService := notice.NewWechatService(webhook)

	a := app.NewApplication(domain.NewAccount(cs), notifyService)
	err = a.Run()
	if err != nil {
		panic(err)
	}
}
