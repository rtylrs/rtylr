package domain

type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IsEnabled   bool   `json:"isEnabled"`
}

// AvailableApps returns the catalog of apps available on the platform.
func AvailableApps() []App {
	return []App{
		{ID: "pos", Name: "Point of Sale", Description: "Manage in-store and online sales", Icon: "shopping-cart", IsEnabled: true},
		{ID: "erp", Name: "Inventory & Supply", Description: "Track inventory, suppliers, and purchase orders", Icon: "package", IsEnabled: true},
		{ID: "hr", Name: "Human Resources", Description: "Manage staff, shifts, payroll, and time off", Icon: "users", IsEnabled: true},
		{ID: "crm", Name: "Customer Relations", Description: "Customer management, loyalty, and campaigns", Icon: "heart", IsEnabled: true},
		{ID: "finance", Name: "Finance", Description: "Invoices, expenses, and tax management", Icon: "dollar-sign", IsEnabled: true},
		{ID: "flow", Name: "Flow", Description: "Projects, tasks, and workspaces", Icon: "activity", IsEnabled: true},
	}
}
