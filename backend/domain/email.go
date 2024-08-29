package domain

import "context"

type MailGunRepo interface{
	SendMass(ctx context.Context, studName string, parentName string, eAddress string) error
}

type MailGunUseCase interface{
	SendMass(ctx context.Context, studName string, parentName string, eAddress string) error
}