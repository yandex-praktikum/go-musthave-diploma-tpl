package domain

type Password string

func (p Password) String() string {
	return string(p)
}
