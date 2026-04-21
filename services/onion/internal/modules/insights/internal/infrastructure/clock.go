package infrastructure

import "time"

// nowFn is an indirection over time.Now so repository-layer range
// computations (e.g. MTD) can be frozen in unit tests without touching
// production code paths. Production always uses time.Now.
var nowFn = time.Now
