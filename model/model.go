package model

type Model struct {
	db
}

func New(db db) *Model {
	return &Model{
		db: db,
	}
}

func (m *Model) People() ([]*User, error) {
	return m.SelectPeople()
}

func (m *Model) CheckLoginUser(login, password string) ([]*User, error) {
	return m.LoginUser(login, password)
}
