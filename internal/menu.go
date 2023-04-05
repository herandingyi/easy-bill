package internal

import (
	"easy-bill/internal/models"
	"xorm.io/xorm"
)

func InsertMenu(s *xorm.Session, m *models.Menu) error {

	_, err := s.Exec("insert into menu"+
		"(name"+
		",code"+
		",price"+
		",currency_type"+
		",supplier) values(?,?,?,?,?) on duplicate key update name = ?,price = ?,currency_type = ?,supplier = ? ", m.Name, m.Code, m.Price, m.CurrencyType, m.Supplier, m.Name, m.Price, m.CurrencyType, m.Supplier)
	if err != nil {
		return err
	}
	return nil
}
