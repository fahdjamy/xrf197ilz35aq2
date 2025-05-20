package postgres

type RowScanError struct {
	Err       error
	SkipCount int64
}

func (e *RowScanError) Error() string {
	return e.Err.Error()
}
