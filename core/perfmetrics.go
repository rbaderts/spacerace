package core

import (
	"time"
)

func TimeCall(start time.Time, name string) {
	elapsed := time.Since(start)
	PerfLog.Printf("%s took %v", name, elapsed)
}
