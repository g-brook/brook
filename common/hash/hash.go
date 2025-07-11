package hash

import (
	"hash/fnv"
	"math"
)

func GetHash32(data string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(data))
	return int(h.Sum32()) % math.MaxInt32
}
