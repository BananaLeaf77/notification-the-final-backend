package usecase

import (
	"context"
	"notification/domain"
	"time"
)

type userUC struct {
	userRepo domain.UserRepo
	TimeOut  time.Duration
}

func NewUserUseCase(repo domain.UserRepo, timeOut time.Duration) domain.UserRepo {
	return &userUC{
		userRepo: repo,
		TimeOut:  timeOut,
	}
}

func (u *userUC) FindUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()
	v, err := u.userRepo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) CreateStaff(ctx context.Context, payload *domain.User) (*domain.User, error) {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()
	v, err := u.userRepo.CreateStaff(ctx, payload)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) GetAllStaff(ctx context.Context) (*[]domain.SafeStaffData, error) {
	// ctx, cancel := context.WithTimeout(ctx, mUC.TimeOut)
	// defer cancel()
	v, err := u.userRepo.GetAllStaff(ctx)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) DeleteStaff(ctx context.Context, id int) error {
	err := u.userRepo.DeleteStaff(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) UpdateStaff(ctx context.Context, id int, payload *domain.User, subjectCodes []string) error {
	err := u.userRepo.UpdateStaff(ctx, id, payload, subjectCodes)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) GetAllTestScoreHistory(ctx context.Context) (*[]domain.TestScore, error) {
	v, err := u.userRepo.GetAllTestScoreHistory(ctx)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) GetStaffDetail(ctx context.Context, id int) (*domain.SafeStaffData, error) {
	v, err := u.userRepo.GetStaffDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) CreateSubject(ctx context.Context, subject *domain.Subject) error {
	err := u.userRepo.CreateSubject(ctx, subject)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) CreateSubjectBulk(ctx context.Context, subjects *[]domain.Subject) (*[]string, error) {
	errList, _ := u.userRepo.CreateSubjectBulk(ctx, subjects)
	if errList != nil {
		return errList, nil
	}

	return nil, nil
}

func (u *userUC) GetAllSubject(ctx context.Context, userID int) (*[]domain.Subject, error) {
	v, err := u.userRepo.GetAllSubject(ctx, userID)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) UpdateSubject(ctx context.Context, subjectCode string, newSubjectData *domain.Subject) error {
	err := u.userRepo.UpdateSubject(ctx, subjectCode, newSubjectData)
	if err != nil {
		return err
	}

	return nil
}

// func (u *userUC) DeleteSubject(ctx context.Context, id int) error {
// 	err := u.userRepo.DeleteSubject(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (u *userUC) GetSubjectsForTeacher(ctx context.Context, userID int) (*domain.SafeStaffData, error) {
	v, err := u.userRepo.GetSubjectsForTeacher(ctx, userID)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) InputTestScores(ctx context.Context, teacherID int, testScores *domain.InputTestScorePayload) error {
	err := u.userRepo.InputTestScores(ctx, teacherID, testScores)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) DeleteStaffMass(ctx context.Context, ids *[]int) error {
	err := u.userRepo.DeleteStaffMass(ctx, ids)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) GetSubjectDetail(ctx context.Context, subjectCode string) (*domain.Subject, error) {
	v, err := u.userRepo.GetSubjectDetail(ctx, subjectCode)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// func (u *userUC) DeleteSubjectMass(ctx context.Context, ids *[]int) error {
// 	err := u.userRepo.DeleteSubjectMass(ctx, ids)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func (u *userUC) GetAllTestScores(ctx context.Context) (*[]domain.TestScore, error) {
	v, err := u.userRepo.GetAllTestScores(ctx)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) GetAllTestScoresBySubjectID(ctx context.Context, subjectCode string) (*[]domain.TestScore, error) {
	v, err := u.userRepo.GetAllTestScoresBySubjectID(ctx, subjectCode)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) GetAdminByAdmin(ctx context.Context) (*domain.SafeStaffData, error) {
	v, err := u.userRepo.GetAdminByAdmin(ctx)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) ShowProfile(ctx context.Context, uID int) (*domain.SafeStaffData, error) {
	v, err := u.userRepo.ShowProfile(ctx, uID)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// func (u *userUC) GetAllAssignedSubject(ctx context.Context, userID int) (*[]domain.Subject, error) {
// 	v, err := u.userRepo.GetAllAssignedSubject(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return v, nil
// }

// func (u *userUC) CreateClass(ctx context.Context, data *domain.Class) error {
// 	err := u.userRepo.CreateClass(ctx, data)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (u *userUC) DeleteClass(ctx context.Context, id int) error {
// 	err := u.userRepo.DeleteClass(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (u *userUC) GetlAllClass(ctx context.Context) (*[]domain.Class, error) {
// 	v, err := u.userRepo.GetlAllClass(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return v, nil
// }
