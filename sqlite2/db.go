package sqlite2

import (
	_ "modernc.org/sqlite"

	"github.com/goflower-io/xsql"
)

func NewDB(c *xsql.Config) (*xsql.DB, error) {
	return xsql.NewDB("sqlite", c)
}
