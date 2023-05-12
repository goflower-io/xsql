package sqlite3

import (
	"github.com/happycrud/xsql"
	_ "github.com/mattn/go-sqlite3"
)

func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("sqlite3", c)
}
