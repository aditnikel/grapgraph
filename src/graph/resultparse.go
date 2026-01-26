package graph

import "fmt"

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
	cols := make([]string, 0, len(headerAny))
	for _, c := range headerAny {
		cols = append(cols, fmt.Sprint(c))
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
		m := map[string]any{}
		for i := 0; i < len(cols) && i < len(rowArr); i++ {
			m[cols[i]] = rowArr[i]
		}
		out = append(out, m)
	}
	return out, nil
}
