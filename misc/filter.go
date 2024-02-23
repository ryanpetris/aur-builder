package misc

func FilterSlice[T any](slice []T, filter func(T) (bool, error)) ([]T, error) {
	var result []T

	for _, item := range slice {
		if keep, err := filter(item); err != nil {
			return nil, err
		} else if !keep {
			continue
		}

		result = append(result, item)
	}

	return result, nil
}

func FilterEmptyString(slice []string) []string {
	result, _ := FilterSlice(slice, func(str string) (bool, error) {
		return str != "", nil
	})

	return result
}
