// Package log defines an implementation-agnostic logging API
// and its implementations on top of `zap` and `zerolog` libraries.
//
// Motivation: since the purpose of this project is learning,
// defining and implementing a logging API serves this purpose well.
// The author is aware that the `log/slog` package in the standard
// library could have been used as an unified logging API frontend,
// with multiple backends for it already available in the wild
// (including `zap` and `zerolog`). There are also other packages
// (e.g. `logr`) that provide an implementation-agnostic logging API.
// Nevertherless, the author has decided to implement his own API
// for the sake of learning.
// The author is also aware of performance implications of the current
// implementation, potentially negating the speed offered by the
// underlying logging libraries, but given the scale of the
// project, not-a-top-notch performance should be acceptable.
// That said, it would be an interesing case to investigate
// when the course reaches the part covering performance optimisation
// with `pprof`.
package log
