package users

type User struct {
	Name     string   `json:"name"`
	ClientID string   `json:"clientId"`
	Position Position `json:"position"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}
