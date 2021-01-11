package integration

type ACMEClient interface {
	FindAccounts() ([]ACMEAccount, error)
	Found()	bool
	String() string
	Name() string
	FindValidationToken() (string, error)
	FindValidationDomain() (string, error)
}
