package common

type Module string

const (
	ModuleBitcoin Module = "bitcoin"
	ModuleRunes   Module = "runes"
)

func (m Module) String() string {
	return string(m)
}
