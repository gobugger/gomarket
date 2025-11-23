package testutil

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Infra struct {
	Db          *pgxpool.Pool
	Mc          *MinioContainer
	terminateDB func()
}

func NewInfra(ctx context.Context) *Infra {
	db, cleanupDB, _ := SetupDB()
	mc, err := NewMinioContainer(ctx)
	if err != nil {
		panic(err)
	}

	return &Infra{db, mc, cleanupDB}
}

func (infra *Infra) CleanDB() error {
	dsn := infra.Db.Config().ConnConfig.ConnString()
	return CleanDB(dsn)
}

func (infra *Infra) Close() error {
	ctx := context.Background()
	infra.terminateDB()
	return infra.Mc.Terminate(ctx)
}
