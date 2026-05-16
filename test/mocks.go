package mocks

import (
	"fmt"

	"github.com/Medzoner/medzoner-go/test/mocks"
	"go.uber.org/mock/gomock"
)

type Mocks struct {
	ContactRepository *mocks.MockRepository
	Mailer            *mocks.MockMailer
	Captcher          *mocks.MockCaptcher
	Validater         *mocks.MockValidater
}

func New(reporter gomock.TestReporter) *Mocks {
	controller := gomock.NewController(reporter)
	fmt.Println(controller)
	return &Mocks{
		ContactRepository: mocks.NewMockRepository(controller),
		Mailer:            mocks.NewMockMailer(controller),
		Captcher:          mocks.NewMockCaptcher(controller),
		Validater:         mocks.NewMockValidater(controller),
	}
}
