package structs

import "bytes"

// generate:reset
type ExternalStructAsField struct {
	buf  *bytes.Buffer
	bufP **bytes.Buffer
}
