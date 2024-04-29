package repository

type ClientStreamRepository struct{}

func NewClientStreamRepository() *ClientStreamRepository {
	return &ClientStreamRepository{}
}

func (r *ClientStreamRepository) GetData() string {
	return "data"
}
