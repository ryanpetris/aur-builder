package pacdb

import (
	"database/sql"
	"errors"
	"reflect"
)

func QueryRow(query string, params []any, args ...any) error {
	if err := connect(); err != nil {
		return err
	}

	if err := database.QueryRow(query, params...).Scan(args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return err
	}

	return nil
}

func getNameMap[T interface{}]() map[string]string {
	var zero [0]T

	t := reflect.TypeOf(zero).Elem()
	fieldCount := t.NumField()
	result := map[string]string{}

	for i := 0; i < fieldCount; i++ {
		field := t.Field(i)

		tag := field.Tag.Get("pacdb")

		if tag == "" {
			tag = field.Name
		}

		result[tag] = field.Name
	}

	return result
}

func QueryStruct[T interface{}](query string, params []any) ([]T, error) {
	if err := connect(); err != nil {
		return nil, err
	}

	rows, err := database.Query(query, params...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var results []T

	nameMap := getNameMap[T]()

	for rows.Next() {
		var result T
		var pointers []any

		if cols, err := rows.Columns(); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}

			return nil, err
		} else {
			for _, colName := range cols {
				pointers = append(pointers, reflect.ValueOf(&result).Elem().FieldByName(nameMap[colName]).Addr().Interface())
			}
		}

		if err := rows.Scan(pointers...); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}

			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}
