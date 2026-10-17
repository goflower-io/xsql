package xsql

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"
)

type ExecQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type DBI interface {
	ExecQuerier
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type Config struct {
	DSN          string        // write data source name.
	ReadDSN      []string      // read data source name.
	Active       int           // pool
	Idle         int           // pool
	IdleTimeout  time.Duration // connect max idle time.
	QueryTimeout time.Duration // query sql timeout
	ExecTimeout  time.Duration // execute sql timeout
}

func NewDB(driver string, c *Config) (*DB, error) {
	m, err := connect(c.DSN, driver, c.Active, c.Idle, c.IdleTimeout)
	if err != nil {
		return nil, err
	}
	var rs []*sql.DB
	if len(c.ReadDSN) == 0 {
		rs = append(rs, m)
	}
	for _, v := range c.ReadDSN {
		r, err := connect(v, driver, c.Active, c.Idle, c.IdleTimeout)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	db := &DB{
		master: m,
		slaves: rs,
		config: c,
	}

	return db, nil
}

func connect(dsn string, driverName string, active, idle int, idleTimeout time.Duration) (*sql.DB, error) {
	m, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	m.SetMaxOpenConns(active)
	m.SetMaxIdleConns(idle)
	m.SetConnMaxIdleTime(idleTimeout)
	return m, nil
}

type DB struct {
	master *sql.DB
	slaves []*sql.DB
	idx    int64
	config *Config
	logger atomic.Pointer[Logger]
}

// SetLogger sets the logger used to record executed SQL. Pass nil to disable
// logging. It is safe to call concurrently with queries.
func (db *DB) SetLogger(l Logger) {
	if l == nil {
		db.logger.Store(nil)
		return
	}
	db.logger.Store(&l)
}

// Logger returns the currently configured logger, or nil when logging is off.
func (db *DB) Logger() Logger {
	if p := db.logger.Load(); p != nil {
		return *p
	}
	return nil
}

func (db *DB) logf(ctx context.Context, format string, v ...any) {
	if l := db.Logger(); l != nil {
		l.Printf(ctx, format, v...)
	}
}

func (db *DB) Config() *Config {
	return db.config
}

func (db *DB) Master() *sql.DB {
	return db.master
}

// slave returns a read replica chosen by round-robin. It returns the master
// when no ReadDSN is configured.
func (db *DB) slave() *sql.DB {
	v := atomic.AddInt64(&db.idx, 1)
	return db.slaves[int(v)%len(db.slaves)]
}

func (db *DB) PingContext(ctx context.Context) error {
	if err := db.master.PingContext(ctx); err != nil {
		return err
	}
	for _, v := range db.slaves {
		if err := v.PingContext(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	r, err := db.master.ExecContext(ctx, query, args...)
	cost := time.Since(start)
	if err != nil {
		db.logf(ctx, "Exec error query:%s args:%+v cost:%s err:%v", query, args, cost, err)
	} else {
		db.logf(ctx, "Exec info  query:%s args:%+v cost:%s", query, args, cost)
	}
	return r, err
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	r, err := db.slave().QueryContext(ctx, query, args...)
	cost := time.Since(start)
	if err != nil {
		db.logf(ctx, "Query error query:%s args:%+v cost:%s err:%v", query, args, cost, err)
	} else {
		db.logf(ctx, "Query info  query:%s args:%+v cost:%s", query, args, cost)
	}
	return r, err
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return db.master.BeginTx(ctx, opts)
}

func (db *DB) Begin() (*sql.Tx, error) {
	return db.master.BeginTx(context.Background(), nil)
}
