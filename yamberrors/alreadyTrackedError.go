package yamberrors

import "fmt"

type AlreadyTrackedError struct {
	Entity string
}

func (e *AlreadyTrackedError) Error() string {
	return fmt.Sprintf("%s is already being tracked.", e.Entity)
}
