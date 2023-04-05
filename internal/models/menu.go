package models

type Menu struct {
	Id           int64
	Code         string
	Name         string
	Supplier     string
	Price        int64
	CurrencyType int
}
