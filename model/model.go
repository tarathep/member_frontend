package model

// Member data
type Member struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Message struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Auth struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	Name string `json:"name"`
}
