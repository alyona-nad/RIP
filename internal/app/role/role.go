package role


type Role int

const (
	Guest   Role = iota // 0
	User             // 1
	Moderator               // 2
)