package database

import "fmt"

type ErrFeedNotFound struct {
	URL string
}

func (e *ErrFeedNotFound) Error() string {
	return fmt.Sprintf("feed not found: %s", e.URL)
}
