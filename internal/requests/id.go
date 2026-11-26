package requests

import (
	"errors"
	"fmt"
	"strings"
	"uuid"
)

// Require dashed UUIDv4 IDs in URL paths:
// https://www.rfc-editor.org/rfc/rfc9562.html#section-5.4
func validateIDs(ids ...string) error {
	for _, id := range ids {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return fmt.Errorf("invalid UUID: %w", err)
		}

		if !strings.EqualFold(id, parsed.String()) {
			return errors.New("UUID must use the dashed format")
		}

		if parsed[6]>>4 != 4 {
			return errors.New("UUID must be version 4")
		}

		if parsed[8]&0xc0 != 0x80 {
			return errors.New("UUID must use the RFC variant")
		}
	}
	return nil
}
