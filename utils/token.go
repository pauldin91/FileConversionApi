package utils

type Generator interface {
	Generate(userId string, username, role string) (string, error)
	Validate(providedToken string) (*CustomClaims, error)
}
