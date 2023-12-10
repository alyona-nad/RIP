package role

/*type Role string

const (
	User   Role = "Пользователь"// 0
	Moderator         Role = "Модератор"
)*/
type Role int

const (
	Guest   Role = iota // 0
	User             // 1
	Moderator               // 2
)