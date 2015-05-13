package vaas

type VaasError struct {
	Msg string
}

func (e VaasError) Error() string {
	return "vaas: " + e.Msg
}
