package domain

type Student struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Class           string `json:"class"`
	Gender          string `json:"gender"`
	TelephoneNumber int64  `json:"telephone_number"`
}

type StudentRepo interface {
	CreateStudent(student *Student) error
	GetAllStudent() (*[]Student, error)
	GetStudentByID(id int) (*Student, error)
	UpdateStudent(id int) error
	DeleteStudent(id int) error
}

type StudentUseCase interface {
	CreateStudentUC(student *Student) error
	GetAllStudentUC() (*[]Student, error)
	GetStudentByIDUC(id int) (*Student, error)
	UpdateStudentUC(id int) error
	DeleteStudentUC(id int) error
}
