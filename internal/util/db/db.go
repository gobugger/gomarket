package db

import (
	"context"
	"fmt"
	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Type registering code is from https://github.com/sqlc-dev/sqlc/issues/2116#issuecomment-1493299852
func Open(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	customTypes, err := getCustomDataTypes(context.Background(), pool)
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, t := range customTypes {
			conn.TypeMap().RegisterType(t)
		}

		pgxdecimal.Register(conn.TypeMap())

		return nil
	}

	// Immediately close the old pool and open a new one with the new config.
	pool.Close()

	pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

// Any custom DB types made with CREATE TYPE need to be registered with pgx.
// https://github.com/kyleconroy/sqlc/issues/2116
// https://stackoverflow.com/questions/75658429/need-to-update-psql-row-of-a-composite-type-in-golang-with-jack-pgx
// https://pkg.go.dev/github.com/jackc/pgx/v5/pgtype
func getCustomDataTypes(ctx context.Context, pool *pgxpool.Pool) ([]*pgtype.Type, error) {
	// Get a single connection just to load type information.
	conn, err := pool.Acquire(ctx)
	defer conn.Release()
	if err != nil {
		return nil, err
	}

	// An underscore prefix is an array type in pgtypes.
	dataTypeNames := []string{
		"invoice_status", "_invoice_status",
		"order_status", "_order_status",
		"dispute_offer_status", "_dispute_offer_status",
		"withdrawal_status", "_withdrawal_status",
	}

	var typesToRegister []*pgtype.Type
	for _, typeName := range dataTypeNames {
		dataType, err := conn.Conn().LoadType(ctx, typeName)
		if err != nil {
			return nil, fmt.Errorf("failed to load type %s: %w", typeName, err)
		}
		// You need to register only for this connection too, otherwise the array type will look for the register element type.
		conn.Conn().TypeMap().RegisterType(dataType)
		typesToRegister = append(typesToRegister, dataType)
	}
	return typesToRegister, nil
}
