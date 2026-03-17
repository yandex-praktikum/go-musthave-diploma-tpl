package domain

type Login string

func (l Login) String() string {
	return string(l)
}
