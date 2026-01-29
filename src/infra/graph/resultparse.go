package graph

import (
	"fmt"
	"strconv"
)

// FalkorDB compact format: [header, rows, stats]
func ParseCompact(resp any) ([]map[string]any, error) {
	arr, ok := resp.([]any)
	if !ok || len(arr) < 2 {
		return nil, fmt.Errorf("unexpected graph response shape")
	}

	headerAny, ok := arr[0].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected header shape")
	}
	cols := make([]string, len(headerAny))
	for i, c := range headerAny {
		cols[i] = cellToString(unpackCell(c))
	}

	rowsAny, ok := arr[1].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected rows shape")
	}

	out := make([]map[string]any, 0, len(rowsAny))
	for _, rowAny := range rowsAny {
		rowArr, ok := rowAny.([]any)
		if !ok {
			continue
		}
		m := make(map[string]any, len(cols))
		for i := 0; i < len(cols) && i < len(rowArr); i++ {
			m[cols[i]] = unpackCell(rowArr[i])
		}
		out = append(out, m)
	}
	return out, nil
}

// Compact response cells are typically [typecode, value].
func unpackCell(v any) any {
	if arr, ok := v.([]any); ok && len(arr) >= 2 {
		return arr[1]
	}
	return v
}

func cellToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case fmt.Stringer:
		return t.String()
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprint(t)
	}
}
