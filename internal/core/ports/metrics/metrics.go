package metrics

//go:generate mockery
type Collector interface {
	IncPVZsCreated()
	IncReceptionsCreated()
	IncProductsAdded()
	IncHTTPRequestsTotal(method, code string)
	ObserveHTTPRequestDuration(method string, duration float64)
}
