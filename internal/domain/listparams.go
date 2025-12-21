package domain

// ListParams represents common pagination parameters for list queries.
// Offset and Limit are sanitized by repositories/usecases to ensure sane bounds.
type ListParams struct {
	Offset int
	Limit  int
}

// Sanitize returns a copy with normalized values.
// - Offset < 0 -> 0
// - Limit <= 0 or > 100 -> 50 (default)
func (p ListParams) Sanitize() ListParams {
	out := p
	if out.Offset < 0 {
		out.Offset = 0
	}
	if out.Limit <= 0 || out.Limit > 100 {
		out.Limit = 50
	}
	return out
}
