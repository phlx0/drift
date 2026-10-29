//go:build !unix

package engine

// setNice is a no-op on platforms without setpriority(2).
func setNice(int) {}
