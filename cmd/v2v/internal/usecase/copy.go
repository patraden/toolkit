package usecase

import (
	"context"

	"github.com/patraden/toolkit/cmd/v2v/internal/domain"
	"github.com/patraden/toolkit/cmd/v2v/internal/repository"
)

type ICopyTableUsecase interface {
	Execute(ctx context.Context, table domain.Table, progressChan chan<- CopyProgress)
}

type Step uint8

const (
	StepGetDDL Step = iota
	StepCreateTable
	StepCopyData
	StepDone
)

func (s Step) String() string {
	return []string{"Get DDL", "Create Table", "Copy Data", "Done"}[s]
}

type CopyProgress struct {
	Step    Step
	Message string
	Error   error
}

type CopyTableUseCase struct {
	sourceRepo repository.TableRepository
	targetRepo repository.TableRepository
}

func NewCopyTableUseCase(sourceRepo, targetRepo repository.TableRepository) *CopyTableUseCase {
	return &CopyTableUseCase{
		sourceRepo: sourceRepo,
		targetRepo: targetRepo,
	}
}

func (uc *CopyTableUseCase) Execute(
	ctx context.Context,
	table domain.Table,
	progressChan chan<- CopyProgress,
) {

	progressChan <- CopyProgress{Step: StepGetDDL, Message: "Getting table DDL from source..."}
	ddl, err := uc.sourceRepo.GetTableDDL(ctx, table)
	if err != nil {
		progressChan <- CopyProgress{
			Step:    StepGetDDL,
			Message: "Failed to get DDL",
			Error:   err,
		}
		return
	}

	progressChan <- CopyProgress{Step: StepCreateTable, Message: "Creating table in target..."}
	if err := uc.targetRepo.CreateTable(ctx, ddl); err != nil {
		progressChan <- CopyProgress{
			Step:    StepCreateTable,
			Message: "Failed to create table",
			Error:   err,
		}
		return
	}

	progressChan <- CopyProgress{Step: StepCopyData, Message: "Copying data..."}
	if err := uc.sourceRepo.CopyData(ctx, table, table); err != nil {
		progressChan <- CopyProgress{
			Step:    StepCopyData,
			Message: "Failed to copy data",
			Error:   err,
		}
		return
	}

	progressChan <- CopyProgress{Step: StepDone, Message: "Copy completed successfully"}
}
