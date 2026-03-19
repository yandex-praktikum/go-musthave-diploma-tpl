package domain

type Token string

func (t Token) String() string {
	return string(t)
}
