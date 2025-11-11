package model

import "context"

type Result struct {
	err error
}

func (r Result) Error() error {
	return r.err
}

type Operator interface {
	Run(ctx context.Context) Result
}

type BackupOperation struct {
	source Source
	volumes []Volume
}

type RestoreOperation struct {
	source Source
	volumes []Volume
}

func (o BackupOperation) Run(ctx context.Context) Result {
	




	return Result{}
}

func (o RestoreOperation) Run(ctx context.Context) Result {
	// TODO: implement restore operation
	return Result{}
}