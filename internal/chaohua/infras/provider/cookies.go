package provider

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
)

type CookieService struct {
	key  []byte
	file string
}

func NewCookieService(keyStr string, file string) (*CookieService, error) {
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, err
	}
	return &CookieService{key: key, file: file}, nil
}

func (c *CookieService) GetSub() (string, error) {
	oldCookies, err := c.OldCookies()
	if err != nil {
		return "", err
	}

	slog.Info("登录中")
	var cookies []*network.Cookie
	for range 3 {
		cookies, err = c.Cookies(oldCookies)
		if err == nil {
			break
		}
	}
	if err != nil {
		slog.Error("登录失败", "err", err)
		return "", err
	}

	for _, cookie := range cookies {
		if cookie.Name == "SUB" && cookie.Domain == ".weibo.com" {
			return cookie.Value, nil
		}
	}
	return "", errors.New("not found")
}

func (c *CookieService) Cookies(oldCookies []*network.CookieParam) (cookies []*network.Cookie, err error) {
	parent, cancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Headless,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
	)
	defer cancel()

	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()

	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, cp := range oldCookies {
				if err := network.SetCookie(cp.Name, cp.Value).WithDomain(cp.Domain).WithPath(cp.Path).WithExpires(cp.Expires).Do(ctx); err != nil {
					return err
				}
			}
			return nil
		}),
		chromedp.Navigate("https://weibo.com"),
		chromedp.WaitVisible("a[href=\"/at/weibo\"][title=\"消息\"]", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			ccs, err := storage.GetCookies().Do(ctx)
			if err != nil {
				return err
			}

			cookies = ccs
			return nil
		}),
	)
	c.SaveCookies(cookies)
	return
}

func (c *CookieService) SaveCookies(cookies []*network.Cookie) error {
	plainText, err := json.Marshal(cookies)
	if err != nil {
		return err
	}

	cipherText, err := c.encrypt(plainText)
	if err != nil {
		return err
	}

	return os.WriteFile(c.file, cipherText, 0644)
}

func (c *CookieService) OldCookies() (oldCookies []*network.CookieParam, err error) {
	cipherText, err := os.ReadFile(c.file)
	if err != nil {
		return nil, err
	}

	plainText, err := c.decrypt(cipherText)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(plainText, &oldCookies)
	return
}

func (c *CookieService) encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func (c *CookieService) decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
