package utils

type Loc struct {
	Country string
	Province string
	Isp      string
}

func NewLoc(ip string)(l *Loc) {
	return new(Loc)
}