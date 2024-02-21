package app

import (
	"context"
	"fmt"

	pg "github.com/go-pg/pg/v10"
)

type dbLogger struct{}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	query, err := q.FormattedQuery()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(query))
	return nil
}
