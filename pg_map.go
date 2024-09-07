package xsql

import (
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
)

var PgxMap = sync.OnceValue(func() *pgtype.Map {
	return pgtype.NewMap()
})
