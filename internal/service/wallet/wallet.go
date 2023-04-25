package wallet

import (
	"database/sql"
	"time"

	"easy-bill/internal/models"

	"xorm.io/xorm"
)

func WalletInc(s *xorm.Session, userId int64, currencyType int, numerator, denominator, commandId int64, createTime time.Time) error {
	if denominator == 0 {
		panic("denominator can not be 0")
	}
	before := &models.Wallet{}
	_, err := s.SQL("select"+
		" id"+
		",user_id"+
		",type"+
		",account_numerator"+
		",account_denominator"+
		" from wallet where user_id = ? and type = ? for update", userId, currencyType).Get(before)
	if err != nil {
		return err
	}
	after := &models.Wallet{}
	if before.Id == 0 {
		var ret sql.Result
		ret, err = s.Exec(`
insert into wallet(user_id,
type,
account_numerator,
account_denominator)
values(?, ?, ?, ?)`, userId, currencyType, numerator, denominator)
		if err != nil {
			return err
		}
		before.AccountNumerator = 0
		before.AccountDenominator = 1
		before.UserId = userId
		before.Type = currencyType

		after.Id, err = ret.LastInsertId()
		if err != nil {
			return err
		}
		after.AccountNumerator = numerator
		after.AccountDenominator = denominator
		after.UserId = userId
		after.Type = currencyType
	} else {
		after.AccountNumerator, after.AccountDenominator = before.Inc(numerator, denominator)
		after.Id = before.Id
		after.UserId = before.UserId
		after.Type = before.Type
		_, err = s.Exec("update wallet set"+
			" account_numerator = ?,"+
			" account_denominator = ?"+
			" where id = ?", after.AccountNumerator,
			after.AccountDenominator,
			after.Id)
	}
	walletLog := &models.WalletLog{
		CommandId:         commandId,
		UserId:            userId,
		Type:              currencyType,
		BeforeNumerator:   before.AccountNumerator,
		BeforeDenominator: before.AccountDenominator,
		Inc:               float64(numerator) / float64(denominator),
		AfterNumerator:    after.AccountNumerator,
		AfterDenominator:  after.AccountDenominator,
		CreateTime:        createTime,
	}
	_, err = s.Table("wallet_log").Insert(walletLog)
	if err != nil {
		return err
	}

	return err
}
