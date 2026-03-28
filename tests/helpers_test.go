package tests

import "github.com/google/uuid"

// nonExistentUUID returns a valid UUID that is guaranteed not to exist in the database.
func nonExistentUUID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}
