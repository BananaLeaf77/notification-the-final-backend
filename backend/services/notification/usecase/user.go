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

func (u *userUC) UpdateStaff(ctx context.Context, id int, payload *domain.User, subjectIDs []int) error {
	err := u.userRepo.UpdateStaff(ctx, id, payload, subjectIDs)
	if err != nil {
		return err
	}

	return nil
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

func (u *userUC) GetAllSubject(ctx context.Context) (*[]domain.Subject, error) {
	v, err := u.userRepo.GetAllSubject(ctx)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) UpdateSubject(ctx context.Context, id int, newSubjectData *domain.Subject) error {
	err := u.userRepo.UpdateSubject(ctx, id, newSubjectData)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) DeleteSubject(ctx context.Context, id int) error {
	err := u.userRepo.DeleteSubject(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) GetSubjectsForTeacher(ctx context.Context, userID int) (*[]domain.Subject, error) {
	v, err := u.userRepo.GetSubjectsForTeacher(ctx, userID)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (u *userUC) InputTestScores(ctx context.Context, teacherID int, testScores *[]domain.TestScore) error {
	err := u.userRepo.InputTestScores(ctx, teacherID, testScores)
	if err != nil {
		return err
	}

	return nil
}

func (u *userUC) DeleteStaffMass(ctx context.Context, ids *[]int) error{
	err := u.userRepo.DeleteStaffMass(ctx, ids)
	if err != nil {
		return err
	}

	return nil
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
