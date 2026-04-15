package uuid

import "github.com/google/uuid"

func ParseUUID(s string) *uuid.UUID {
	p, err := uuid.Parse(s)
	if err != nil {
		return nil
	}

	return &p
}
