package app

import "github.com/joejacegale/chaohua/internal/chaohua/domain"

type Application struct {
	account *domain.Account
	notice  NotifyService
}

func NewApplication(account *domain.Account, notice NotifyService) *Application {
	return &Application{
		account: account,
		notice:  notice,
	}
}

func (a *Application) Run() error {
	err := a.account.Exec()
	if err != nil {
		return err
	}
	return a.notice.Notify(a.account.EventMessage())
}

type NotifyService interface {
	Notify(content string) error
}
