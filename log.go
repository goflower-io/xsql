package xsql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"
)

// Logger records executed SQL statements and errors. The ctx argument carries
// the caller's context so logger implementations can propagate it (e.g. to
// slog.DebugContext). Use stdLogger / SlogLogger to adapt common loggers.
type Logger interface {
	Printf(ctx context.Context, format string, v ...any)
}

// StdLogger adapts a *log.Logger (e.g. log.Default()) to Logger, dropping ctx.
func StdLogger(l *log.Logger) Logger {
	if l == nil {
		l = log.Default()
	}
	return stdLogger{l}
}

type stdLogger struct{ l *log.Logger }

func (s stdLogger) Printf(ctx context.Context, format string, v ...any) {
	s.l.Printf(format, v...)
}

// DebugOption configures the DebugDB returned by Debug.
type DebugOption func(*DebugDB)

// WithLogger overrides the default logger (log.Default()); nil keeps it.
func WithLogger(l Logger) DebugOption {
	return func(d *DebugDB) {
		if l != nil {
			d.log = l
		}
	}
}

// Debug wraps db and logs every executed SQL statement and error. It defaults
// to log.Default(); use WithLogger to route output elsewhere.
func Debug(db DBI, opts ...DebugOption) *DebugDB {
	d := &DebugDB{dbt: db, log: StdLogger(log.Default())}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// SlogLogger adapts a *slog.Logger to Logger using the debug level.
func SlogLogger(l *slog.Logger) Logger {
	if l == nil {
		l = slog.Default()
	}
	return slogLogger{l}
}

type slogLogger struct{ l *slog.Logger }

func (s slogLogger) Printf(ctx context.Context, format string, v ...any) {
	s.l.DebugContext(ctx, fmt.Sprintf(format, v...))
}

type DebugDB struct {
	log Logger
	dbt DBI
}

type DebugTx struct {
	log Logger
	eq  ExecQuerier
}

func (d *DebugDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	r, err := d.dbt.ExecContext(ctx, query, args...)
	cost := time.Since(start)
	if err != nil {
		d.log.Printf(ctx, "Debug Exec error query:%s args:%+v cost:%s err:%v", query, args, cost, err)
	} else {
		d.log.Printf(ctx, "Debug Exec query:%s args:%+v cost:%s", query, args, cost)
	}
	return r, err
}

func (d *DebugDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	r, err := d.dbt.QueryContext(ctx, query, args...)
	cost := time.Since(start)
	if err != nil {
		d.log.Printf(ctx, "Debug Query error query:%s args:%+v cost:%s err:%v", query, args, cost, err)
	} else {
		d.log.Printf(ctx, "Debug Query query:%s args:%+v cost:%s", query, args, cost)
	}
	return r, err
}

func (d *DebugDB) Begin() (*DebugTx, error) {
	tx, err := d.dbt.Begin()
	if err != nil {
		return nil, err
	}
	return &DebugTx{eq: tx, log: d.log}, nil
}

func (d *DebugDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*DebugTx, error) {
	tx, err := d.dbt.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &DebugTx{eq: tx, log: d.log}, nil
}

func (d *DebugTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	r, err := d.eq.ExecContext(ctx, query, args...)
	cost := time.Since(start)
	if err != nil {
		d.log.Printf(ctx, "Debug Tx Exec error query:%s args:%+v cost:%s err:%v", query, args, cost, err)
	} else {
		d.log.Printf(ctx, "Debug Tx Exec query:%s args:%+v cost:%s", query, args, cost)
	}
	return r, err
}

func (d *DebugTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	r, err := d.eq.QueryContext(ctx, query, args...)
	cost := time.Since(start)
	if err != nil {
		d.log.Printf(ctx, "Debug TX Query error query:%s args:%+v cost:%s err:%v", query, args, cost, err)
	} else {
		d.log.Printf(ctx, "Debug TX Query query:%s args:%+v cost:%s", query, args, cost)
	}
	return r, err
}
