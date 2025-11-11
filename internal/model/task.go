package model

import (
	"context"
	"errors"
)

type operation int

const (
	typeBackup operation = iota + 1
	typeRestore
)

type Task interface {
	RunOperation(ctx context.Context) Result
}

type task struct {
	operation
	name    string
	filter  []Filter
	source  []Source
	volumes []Volume
}

func (t task) RunOperation(ctx context.Context, source Source) Result {
	switch t.operation {
	case typeBackup:
		return t.backup(source).Run(ctx)
	case typeRestore:
		return t.restore(source).Run(ctx)
	default:
		return Result{err: errors.New("unknown operation")}
	}
}

func (t task) backup(source Source) Operator {
	return BackupOperation{
		source:  source,
		volumes: t.volumes,
	}
}

func (t task) restore(source Source) Operator {
	return RestoreOperation{
		source:  source,
		volumes: t.volumes,
	}
}
