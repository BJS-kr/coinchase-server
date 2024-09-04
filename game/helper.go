package game

import "math/rand/v2"

func GenerateRandomDirection() int32 {
	if rand.Int32N(2) == 0 {
		return -1
	}

	return 1
}
