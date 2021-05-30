package pg

import (
	"context"
	"encoding/json"
	"explorer"
	"fmt"
	"reflect"

	"github.com/doug-martin/goqu/v9"
)

type query struct {
	Body         string
	DestElemType reflect.Type
}

var queries = map[string]query{}

type Query struct {
	Name     string
	Body     string
	DestElem interface{}
}

// RegisterQuery registers DB query. Panics if `destElem` not struct.
// `destElem` will be used to construct query result slice.
func RegisterQuery(q Query) {
	t := reflect.TypeOf(q.DestElem)

	if !(t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr &&
		t.Elem().Kind() == reflect.Struct)) {
		panic(fmt.Sprintf("expected struct but got %s", t.Kind()))
	}

	queries[q.Name] = query{
		Body:         q.Body,
		DestElemType: t,
	}
}

// RunQuery runs query with given name and JSON encoded args array and returns
// JSON encoded result data
func RunQuery(ctx context.Context, db *goqu.Database, queryName string,
	queryArgs string) (string, error) {

	q, exists := queries[queryName]
	if !exists {
		return "", fmt.Errorf("query `%s` does not exists", queryName)
	}

	var args []interface{}

	if queryArgs != "" {
		err := json.Unmarshal([]byte(queryArgs), &args)
		if err != nil {
			return "", err
		}
	}

	dest := reflect.New(reflect.SliceOf(q.DestElemType))

	fmt.Println(args)

	err := db.ScanStructsContext(ctx, dest.Interface(), q.Body, args...)
	if err != nil {
		return "", err
	}

	destJSON, err := json.Marshal(dest.Interface())
	if err != nil {
		return "", err
	}

	return string(destJSON), nil
}

func (e *Explorer) GetQuery(ctx context.Context, req *explorer.GetQueryReq) (
	*explorer.GetQueryRes, error) {

	data, err := RunQuery(ctx, e.db, req.Name, req.Args)
	if err != nil {
		return nil, err
	}

	return &explorer.GetQueryRes{
		Data: data,
	}, nil
}
