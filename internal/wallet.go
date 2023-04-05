package internal

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
		ret, err = s.Exec("insert into wallet(user_id,"+
			"type,"+
			"account_numerator,"+
			"account_denominator)"+
			" values(?, ?, ?, ?)", userId, currencyType, numerator, denominator)
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

func WalletRollback(s *xorm.Session, uid int, oCid int64, nCid int64) (int, error) {
	var currencyType int

	//查询出上一次操作的日志
	var walletLogs []*models.WalletLog
	err := s.SQL("SELECT type,"+
		"inc,"+
		"user_id,"+
		"before_numerator,"+
		"before_denominator,"+
		"after_numerator,"+
		"after_denominator"+
		" FROM wallet_log  WHERE command_id = ?", oCid).Find(&walletLogs)
	if err != nil {
		return 0, err
	}
	//新增回滚日志
	for _, log := range walletLogs {
		currencyType = log.Type
		walletLog := &models.WalletLog{
			CommandId:         nCid,
			UserId:            int64(uid),
			Type:              log.Type,
			BeforeNumerator:   log.AfterNumerator,
			BeforeDenominator: log.AfterDenominator,
			Inc:               -log.Inc,
			AfterNumerator:    log.BeforeNumerator,
			AfterDenominator:  log.BeforeDenominator,
			CreateTime:        time.Now(),
		}
		_, err = s.Table("wallet_log").Insert(walletLog)
		if err != nil {
			return 0, err
		}
		//钱包金额回滚
		before := &models.Wallet{}
		_, err = s.SQL("select"+
			" id"+
			" from wallet where user_id = ? and type = ? for update", uid, currencyType).Get(before)
		if err != nil {
			return 0, err
		}

		_, err = s.Exec("update wallet set"+
			" account_numerator = ?,"+
			" account_denominator = ?"+
			" where user_id = ?", log.BeforeNumerator,
			log.BeforeDenominator,
			log.UserId)
		if err != nil {
			return 0, err
		}
	}
	return currencyType, nil
}
