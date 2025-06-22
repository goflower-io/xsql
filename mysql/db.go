package mysql

import (
	_ "github.com/go-sql-driver/mysql"

	"github.com/goflower-io/xsql"
)

func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("mysql", c)
}
