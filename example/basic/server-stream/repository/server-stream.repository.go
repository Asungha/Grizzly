package repository

type ServerStreamRepository struct{}

func NewServerStreamRepository() *ServerStreamRepository {
	return &ServerStreamRepository{}
}

func (r *ServerStreamRepository) GetData() string {
	return "stream data"
}
