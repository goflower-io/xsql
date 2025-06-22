package sqlite3

import (
	_ "github.com/mattn/go-sqlite3"

	"github.com/goflower-io/xsql"
)

func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("sqlite3", c)
}
