package infrastructure

type featureRecord struct {
	Tables         bool `json:"tables"`
	KitchenDisplay bool `json:"kitchen_display"`
	Modifiers      bool `json:"modifiers"`
	Combos         bool `json:"combos"`
	Tips           bool `json:"tips"`
	GiftCards      bool `json:"gift_cards"`
	Customers      bool `json:"customers"`
	Inventory      bool `json:"inventory"`
	CashDrawer     bool `json:"cash_drawer"`
}
