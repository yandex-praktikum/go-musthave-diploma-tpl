package domain

type UserAgent string

func (ua UserAgent) String() string {
	return string(ua)
}
