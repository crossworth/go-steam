package protocol

import (
	"math"
	"strconv"
)

type JobID uint64

func (j JobID) String() string {
	if j == math.MaxUint64 {
		return "(none)"
	}
	return strconv.FormatUint(uint64(j), 10)
}
