package repository

type UnaryRepository struct{}

func NewUnaryRepository() *UnaryRepository {
	return &UnaryRepository{}
}

func (r *UnaryRepository) GetData() string {
	return "data"
}
