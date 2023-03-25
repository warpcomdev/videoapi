package auth

type Error string

func (err Error) Error() string {
	return string(err)
}

const (
	ErrorMisingAuthHeader        Error = "missing authorization header"
	ErrorInvalidAuthHeader       Error = "invalid authorization header"
	ErrorUnexpectedSigningMethod Error = "unexpected signing method"
	ErrorInvalidToken            Error = "invalid token"
	ErrorInvalidRole             Error = "invalid role"
	ErrorMissingRole             Error = "missing role"
)
