package app

import (
	"github.com/alexedwards/scs/v2"
	"github.com/gobugger/gomarket/internal/util/uow"
	"github.com/gobugger/gomarket/pkg/payment/processor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/riverqueue/river"
)

type Application struct {
	Db               *pgxpool.Pool
	SessionManager   *scs.SessionManager
	MinioClient      *minio.Client
	PaymentProcessor processor.Processor
	RiverClient      *river.Client[pgx.Tx]
	uow.UoW
}
