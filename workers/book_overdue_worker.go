package workers

import (
	service "github.com/dutt23/lms/services"
)

type finesOverdueWorker struct {
	loanService *service.LoanService
}

func NewFinesOverdueWorker(loanService *service.LoanService) finesOverdueWorker {
	return finesOverdueWorker{
		loanService,
	}
}

func (worker *finesOverdueWorker) OverdueFinesWorker() {}
