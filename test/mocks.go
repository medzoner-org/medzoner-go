package mocks

import (
	"fmt"

	"github.com/Medzoner/medzoner-go/test/mocks"
	"go.uber.org/mock/gomock"
)

type Mocks struct {
	ContactRepository *mocks.MockContactRepository
	TechnoRepository  *mocks.MockTechnoRepository
	Mailer            *mocks.MockMailer
	Captcher          *mocks.MockCaptcher
	Validater         *mocks.MockValidater
}

func New(reporter gomock.TestReporter) *Mocks {
	controller := gomock.NewController(reporter)
	fmt.Println(controller)
	return &Mocks{
		ContactRepository: mocks.NewMockContactRepository(controller),
		Mailer:            mocks.NewMockMailer(controller),
		TechnoRepository:  mocks.NewMockTechnoRepository(controller),
		Captcher:          mocks.NewMockCaptcher(controller),
		Validater:         mocks.NewMockValidater(controller),
	}
}
