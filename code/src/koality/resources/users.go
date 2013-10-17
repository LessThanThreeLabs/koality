package resources

type User struct {
	Id        int
	Email     string
	FirstName string
	LastName  string
}

type UsersHandler struct {
	Read UsersReadHandler
	// Update UsersUpdateHandler
}

type UsersReadHandler interface {
	Get(userId int) (*User, error)
	// GetFromEmail(email string) (User, error)
	// GetAll() ([]User, error)
}

// type UsersUpdateHandler interface {
// 	SetName(id int, firstName, lastName string) error
// }
