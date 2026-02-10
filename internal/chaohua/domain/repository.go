package domain

type Provider interface {
	ChaoHuaList() ([]*ChaoHua, error)
	Sign(ch *ChaoHua) error
}
