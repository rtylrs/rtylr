package crm

// CustomerContact is the name and email Insights shows for a customer.
type CustomerContact struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
}
