package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"

	drivermysql "github.com/go-sql-driver/mysql"
)

const mysqlCompatDriverName = "axisrelay-mysql"

func init() {
	sql.Register(mysqlCompatDriverName, &mysqlCompatDriver{})
}

type mysqlCompatDriver struct {
	base drivermysql.MySQLDriver
}

func (d *mysqlCompatDriver) Open(dsn string) (driver.Conn, error) {
	conn, err := d.base.Open(dsn)
	if err != nil {
		return nil, err
	}
	return &mysqlCompatConn{Conn: conn}, nil
}

func (d *mysqlCompatDriver) OpenConnector(dsn string) (driver.Connector, error) {
	connector, err := d.base.OpenConnector(dsn)
	if err != nil {
		return nil, err
	}
	return &mysqlCompatConnector{base: connector, driver: d}, nil
}

type mysqlCompatConnector struct {
	base   driver.Connector
	driver driver.Driver
}

func (c *mysqlCompatConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.base.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &mysqlCompatConn{Conn: conn}, nil
}

func (c *mysqlCompatConnector) Driver() driver.Driver {
	return c.driver
}

type mysqlCompatConn struct {
	driver.Conn
}

func (c *mysqlCompatConn) Prepare(query string) (driver.Stmt, error) {
	rewritten, order, err := rewriteMySQLPlaceholders(query)
	if err != nil {
		return nil, err
	}
	stmt, err := c.Conn.Prepare(rewritten)
	if err != nil {
		return nil, err
	}
	return &mysqlCompatStmt{Stmt: stmt, order: order}, nil
}

func (c *mysqlCompatConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	rewritten, order, err := rewriteMySQLPlaceholders(query)
	if err != nil {
		return nil, err
	}
	if preparer, ok := c.Conn.(driver.ConnPrepareContext); ok {
		stmt, err := preparer.PrepareContext(ctx, rewritten)
		if err != nil {
			return nil, err
		}
		return &mysqlCompatStmt{Stmt: stmt, order: order}, nil
	}
	return c.Prepare(query)
}

func (c *mysqlCompatConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	rewritten, order, err := rewriteMySQLPlaceholders(query)
	if err != nil {
		return nil, err
	}
	expanded, err := expandMySQLNamedValues(order, args)
	if err != nil {
		return nil, err
	}
	return execer.ExecContext(ctx, rewritten, expanded)
}

func (c *mysqlCompatConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	rewritten, order, err := rewriteMySQLPlaceholders(query)
	if err != nil {
		return nil, err
	}
	expanded, err := expandMySQLNamedValues(order, args)
	if err != nil {
		return nil, err
	}
	return queryer.QueryContext(ctx, rewritten, expanded)
}

func (c *mysqlCompatConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if beginner, ok := c.Conn.(driver.ConnBeginTx); ok {
		return beginner.BeginTx(ctx, opts)
	}
	return c.Conn.Begin()
}

func (c *mysqlCompatConn) Ping(ctx context.Context) error {
	if pinger, ok := c.Conn.(driver.Pinger); ok {
		return pinger.Ping(ctx)
	}
	return nil
}

func (c *mysqlCompatConn) CheckNamedValue(value *driver.NamedValue) error {
	if checker, ok := c.Conn.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(value)
	}
	return driver.ErrSkip
}

func (c *mysqlCompatConn) ResetSession(ctx context.Context) error {
	if resetter, ok := c.Conn.(driver.SessionResetter); ok {
		return resetter.ResetSession(ctx)
	}
	return nil
}

func (c *mysqlCompatConn) IsValid() bool {
	if validator, ok := c.Conn.(driver.Validator); ok {
		return validator.IsValid()
	}
	return true
}

type mysqlCompatStmt struct {
	driver.Stmt
	order []int
}

func (s *mysqlCompatStmt) NumInput() int {
	if len(s.order) == 0 {
		return s.Stmt.NumInput()
	}
	return -1
}

func (s *mysqlCompatStmt) Exec(args []driver.Value) (driver.Result, error) {
	expanded, err := expandMySQLValues(s.order, args)
	if err != nil {
		return nil, err
	}
	return s.Stmt.Exec(expanded)
}

func (s *mysqlCompatStmt) Query(args []driver.Value) (driver.Rows, error) {
	expanded, err := expandMySQLValues(s.order, args)
	if err != nil {
		return nil, err
	}
	return s.Stmt.Query(expanded)
}

func (s *mysqlCompatStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := s.Stmt.(driver.StmtExecContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	expanded, err := expandMySQLNamedValues(s.order, args)
	if err != nil {
		return nil, err
	}
	return execer.ExecContext(ctx, expanded)
}

func (s *mysqlCompatStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := s.Stmt.(driver.StmtQueryContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	expanded, err := expandMySQLNamedValues(s.order, args)
	if err != nil {
		return nil, err
	}
	return queryer.QueryContext(ctx, expanded)
}

func rewriteMySQLPlaceholders(query string) (string, []int, error) {
	order := make([]int, 0, 8)
	out := make([]byte, 0, len(query))
	var quote byte
	lineComment := false
	blockComment := false

	for i := 0; i < len(query); i++ {
		ch := query[i]
		if lineComment {
			out = append(out, ch)
			if ch == '\n' {
				lineComment = false
			}
			continue
		}
		if blockComment {
			out = append(out, ch)
			if ch == '*' && i+1 < len(query) && query[i+1] == '/' {
				out = append(out, '/')
				i++
				blockComment = false
			}
			continue
		}
		if quote != 0 {
			out = append(out, ch)
			if ch == '\\' && i+1 < len(query) {
				out = append(out, query[i+1])
				i++
				continue
			}
			if ch == quote {
				if i+1 < len(query) && query[i+1] == quote {
					out = append(out, query[i+1])
					i++
					continue
				}
				quote = 0
			}
			continue
		}
		if ch == '-' && i+1 < len(query) && query[i+1] == '-' {
			out = append(out, '-', '-')
			i++
			lineComment = true
			continue
		}
		if ch == '#' {
			out = append(out, ch)
			lineComment = true
			continue
		}
		if ch == '/' && i+1 < len(query) && query[i+1] == '*' {
			out = append(out, '/', '*')
			i++
			blockComment = true
			continue
		}
		if ch == '\'' || ch == '"' || ch == 0x60 {
			quote = ch
			out = append(out, ch)
			continue
		}
		if ch == '$' && i+1 < len(query) && query[i+1] >= '1' && query[i+1] <= '9' {
			j := i + 1
			for j < len(query) && query[j] >= '0' && query[j] <= '9' {
				j++
			}
			index, err := strconv.Atoi(query[i+1 : j])
			if err != nil || index <= 0 {
				return "", nil, fmt.Errorf("invalid SQL placeholder %q", query[i:j])
			}
			order = append(order, index)
			out = append(out, '?')
			i = j - 1
			continue
		}
		out = append(out, ch)
	}
	return string(out), order, nil
}

func expandMySQLNamedValues(order []int, args []driver.NamedValue) ([]driver.NamedValue, error) {
	if len(order) == 0 {
		return args, nil
	}
	expanded := make([]driver.NamedValue, len(order))
	for i, index := range order {
		if index < 1 || index > len(args) {
			return nil, fmt.Errorf("SQL placeholder $%d has no argument (got %d)", index, len(args))
		}
		value := args[index-1]
		value.Ordinal = i + 1
		value.Name = ""
		expanded[i] = value
	}
	return expanded, nil
}

func expandMySQLValues(order []int, args []driver.Value) ([]driver.Value, error) {
	if len(order) == 0 {
		return args, nil
	}
	expanded := make([]driver.Value, len(order))
	for i, index := range order {
		if index < 1 || index > len(args) {
			return nil, fmt.Errorf("SQL placeholder $%d has no argument (got %d)", index, len(args))
		}
		expanded[i] = args[index-1]
	}
	return expanded, nil
}
