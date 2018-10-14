package model

type db interface {
	SelectPeople() ([]*User, error)
	LoginUser(login, password string) ([]*User, error)
}
