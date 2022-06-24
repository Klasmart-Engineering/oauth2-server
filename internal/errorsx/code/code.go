package code

type Code string

const (
	NOT_FOUND Code = "NOT_FOUND"
	// NB: May need to refine this in future, although in theory the API gateway should handle specific
	// cases such as 'too short/long', 'missing required', 'invalid type' etc.
	INVALID_ARGUMENT = "INVALID_ARGUMENT"
	INVALID_METHOD   = "INVALID_METHOD"
	REQUIRED_HEADER  = "REQUIRED_HEADER"
	INTERNAL         = "INTERNAL"
)
