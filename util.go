package main

import "errors"

func Pop[T any](s []T) ([]T, T, error)  {
	if len(s) < 1 {
		var zeroT T
		return s, zeroT, errors.New("slice is empty")
	}

	popped := s[len(s) - 1]
	newSlice := s[:len(s) - 1]
	
	return newSlice, popped, nil
}