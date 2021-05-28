package pg

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"
)

type query struct {
	Body         string
	DestElemType reflect.Type
}

var queries = map[string]query{}

// RegisterQuery registers DB query. Panics if `destElem` not struct.
// `destElem` will be used to construct query result slice.
func RegisterQuery(name, body string, destElem interface{}) {
	t := reflect.TypeOf(destElem)

	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected struct but got %s", t.Kind()))
	}

	queries[name] = query{
		Body:         body,
		DestElemType: t,
	}
}

// Select performs select query and returns result slice with type corresponding
// to the query.
func Select(ctx context.Context, db *sqlx.DB, queryName string,
	args []interface{}) (interface{}, error) {

	q, exists := queries[queryName]
	if !exists {
		return nil, fmt.Errorf("query `%s` does not exists", queryName)
	}

	dest := reflect.MakeSlice(q.DestElemType, 0, 0)

	err := db.SelectContext(ctx, dest.Addr().Interface(), q.Body, args...)
	if err != nil {
		return nil, err
	}

	return dest.Interface(), nil
}
