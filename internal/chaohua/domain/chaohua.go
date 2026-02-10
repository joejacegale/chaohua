package domain

import "fmt"

type ChaoHua struct {
	id    string
	title string
}

func (ch *ChaoHua) ID() string {
	return ch.id
}

func (ch *ChaoHua) Title() string {
	return ch.title
}

func NewChaoHua(id, title string) *ChaoHua {
	return &ChaoHua{
		id:    id,
		title: title,
	}
}

type Event struct {
	id    string
	title string
	err   error
}

type Account struct {
	repo   Provider
	events []Event
}

func NewAccount(srv Provider) *Account {
	return &Account{
		repo: srv,
	}
}

func (a *Account) Exec() error {
	chs, err := a.repo.ChaoHuaList()
	if err != nil {
		return err
	}
	for _, ch := range chs {
		err = a.repo.Sign(ch)
		a.events = append(a.events, Event{id: ch.id, title: ch.title, err: err})
	}
	return nil
}

func (a *Account) Events() []Event {
	return a.events
}

func (a *Account) EventMessage() string {
	total := len(a.events)
	failed := 0
	for _, e := range a.events {
		if e.err != nil {
			failed++
		}
	}
	return fmt.Sprintf(
		"一共 %d 个超话\n签到成功 %d 个\n签到失败 %d 个",
		total,
		total-failed,
		failed,
	)
}
