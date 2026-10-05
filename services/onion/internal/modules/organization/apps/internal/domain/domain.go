package domain

type App struct {
	ID          string
	Name        string
	Description string
	Icon        string
	IsEnabled   bool
	// AddonGated indicates this app's visibility is controlled by an addon.
	// If empty, the app is always visible (core app).
	AddonGated string
}

// AvailableApps returns the catalog of apps available on the platform.
// All listed apps are core apps and always enabled.
func AvailableApps() []App {
	return []App{
		{ID: "pos", Name: "Point of Sale", Description: "Manage in-store and online sales", Icon: "shopping-cart", IsEnabled: true},
		{ID: "erp", Name: "Inventory & Supply", Description: "Track inventory, suppliers, and purchase orders", Icon: "package", IsEnabled: true},
		{ID: "hr", Name: "Human Resources", Description: "Manage staff, shifts, payroll, and time off", Icon: "users", IsEnabled: true},
		{ID: "crm", Name: "Customer Relations", Description: "Customer management, loyalty, and campaigns", Icon: "heart", IsEnabled: true},
		{ID: "finance", Name: "Finance", Description: "Invoices, expenses, and tax management", Icon: "dollar-sign", IsEnabled: true},
		{ID: "flow", Name: "Flow", Description: "Projects, tasks, and workspaces", Icon: "activity", IsEnabled: true},
		{ID: "insights", Name: "Insights", Description: "Analytics, reports, and business intelligence", Icon: "bar-chart", IsEnabled: true},
	}
}
