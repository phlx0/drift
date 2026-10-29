//go:build !unix

package engine

import "os"

func claimRender()   {}
func releaseRender() {}

// pauseSignal is nil here: no cross-instance arbitration without SIGUSR1.
var pauseSignal os.Signal
