package domain

type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IsEnabled   bool   `json:"isEnabled"`
	// AddonGated indicates this app's visibility is controlled by an addon.
	// If empty, the app is always visible (core app).
	AddonGated string `json:"-"`
}

// AvailableApps returns the catalog of apps available on the platform.
// Core apps (pos, erp, hr, crm, finance, flow, docs) are always enabled.
// Addon-gated apps check the business_addons table for enablement.
func AvailableApps() []App {
	return []App{
		// Core apps - always enabled
		{ID: "pos", Name: "Point of Sale", Description: "Manage in-store and online sales", Icon: "shopping-cart", IsEnabled: true},
		{ID: "erp", Name: "Inventory & Supply", Description: "Track inventory, suppliers, and purchase orders", Icon: "package", IsEnabled: true},
		{ID: "hr", Name: "Human Resources", Description: "Manage staff, shifts, payroll, and time off", Icon: "users", IsEnabled: true},
		{ID: "crm", Name: "Customer Relations", Description: "Customer management, loyalty, and campaigns", Icon: "heart", IsEnabled: true},
		{ID: "finance", Name: "Finance", Description: "Invoices, expenses, and tax management", Icon: "dollar-sign", IsEnabled: true},
		{ID: "flow", Name: "Flow", Description: "Projects, tasks, and workspaces", Icon: "activity", IsEnabled: true},
		// Addon-gated apps - visibility controlled by business_addons
		{ID: "store", Name: "Store", Description: "Online storefront, products, and orders", Icon: "shopping-bag", IsEnabled: true, AddonGated: "store"},
		{ID: "book", Name: "Book", Description: "Appointments, scheduling, and reservations", Icon: "calendar-check", IsEnabled: true, AddonGated: "book"},
		{ID: "support", Name: "Support", Description: "Help desk, tickets, and live chat", Icon: "headphones", IsEnabled: true, AddonGated: "support"},
		{ID: "insights", Name: "Insights", Description: "Analytics, reports, and business intelligence", Icon: "bar-chart", IsEnabled: true, AddonGated: "insights"},
	}
}
