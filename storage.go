package kexas

import "github.com/jankylewis/kexas/kcore"

// Storage is an alias for kcore.Storage — they are the same type.
// Used for both localStorage and sessionStorage, distinguished by StorageType.
type Storage = kcore.Storage

// StorageType identifies which storage scope a Storage instance operates on.
type StorageType = kcore.StorageType

// StorageType values.
const (
	StorageTypeLocal   = kcore.StorageTypeLocal
	StorageTypeSession = kcore.StorageTypeSession
)
