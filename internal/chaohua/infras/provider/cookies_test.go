package provider_test

import (
	"os"
	"testing"

	repository "github.com/joejacegale/chaohua/internal/chaohua/infras/provider"
)

func TestCookies(t *testing.T) {
	keyStr := os.Getenv("COOKIE_SECRET")
	file := os.Getenv("COOKIE_FILE")

	cs, err := repository.NewCookieService(keyStr, file)
	if err != nil {
		t.Fatal(err)
	}
	cookies, err := cs.Cookies(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cookies)
}
