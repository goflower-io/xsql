package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/goflower-io/xsql"
)

func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("pgx", c)
}
