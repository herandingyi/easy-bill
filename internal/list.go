package internal

import (
	"easy-bill/internal/album"
	"easy-bill/internal/models"
	"fmt"
	"github.com/tucnak/telebot"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"xorm.io/xorm"
)

func GroupNameList(db *xorm.Engine) (a telebot.Album, err error) {
	users := make([]*models.User, 0, 5)
	err = db.SQL("SELECT " +
		" `name`," +
		" `group_name`" +
		" FROM user").Find(&users)
	if err != nil {
		return nil, err
	}
	body := make([][]string, 0, 5)
	for _, user := range users {
		body = append(body, []string{user.Name, user.GroupName})
	}
	return album.ToAlbum([]string{"NAME", "GROUP-NAME"}, body)
}

func MenuList(db *xorm.Engine) (telebot.Album, error) {
	type Menu struct {
		Code     string
		Name     string
		Price    string
		Supplier string
		Currency string
	}
	var err error
	menus := make([]*Menu, 0, 5)
	var sql string
	currencyType := 1
	decimalPlace := fmt.Sprint(CurrencyDecimalPlace[currencyType])
	multi := fmt.Sprint(int(math.Pow(10, float64(CurrencyDecimalPlace[currencyType]))))
	sql = "ROUND( price / " + multi + ", " + decimalPlace + " )price,"
	err = db.SQL("SELECT " +
		"`code`," +
		" `name`," +
		sql +
		"supplier," +
		"IF(currency_type = 1,'$','៛') currency FROM menu WHERE menu.currency_type = " + strconv.Itoa(currencyType)).Find(&menus)
	if err != nil {
		return nil, err
	}
	body := make([][]string, 0, 5)
	for _, menu := range menus {

		body = append(body, []string{menu.Code, menu.Name, menu.Currency + menu.Price, menu.Supplier})
	}
	return album.ToAlbum([]string{"code", "name", "price", "supplier"}, body)
}

func WalletList(db *xorm.Engine, currencyType int, detail bool) (telebot.Album, error) {
	type Wallet struct {
		Name               string
		Remain             string
		AccountNumerator   int64
		AccountDenominator int64
	}
	var err error
	wallets := make([]*Wallet, 0, 5)
	var sql string
	decimalPlace := fmt.Sprint(CurrencyDecimalPlace[currencyType])
	multiInt := int(math.Pow(10, float64(CurrencyDecimalPlace[currencyType])))
	multi := fmt.Sprint(multiInt)
	sql = "CASE WHEN ROUND(ifnull( wallet.account_numerator, 0 )/ " + multi +
		"/ifnull( wallet.account_denominator, 1 ), " + decimalPlace +
		") = ifnull( wallet.account_numerator, 0 )/ " + multi +
		"/ifnull( wallet.account_denominator, 1 )   THEN ROUND(ifnull( wallet.account_numerator, 0 )/ " + multi +
		"/ifnull( wallet.account_denominator, 1 ), " + decimalPlace +
		")    ELSE CONCAT(CONVERT(ifnull( wallet.account_numerator, 0 )/ " + multi +
		",DECIMAL(15," + decimalPlace +
		")), '/', ifnull( wallet.account_denominator, 1 )) END AS remain"

	err = db.SQL("select" +
		" user.name," +
		" wallet.account_numerator," +
		" wallet.account_denominator," +
		sql +
		" from user" +
		" left join wallet on wallet.user_id=user.id and wallet.type=" + strconv.Itoa(currencyType) +
		" where user.status = 1").Find(&wallets)
	if err != nil {
		return nil, err
	}
	body := make([][]string, 0, 5)
	for _, wallet := range wallets {
		if strings.HasSuffix(wallet.Remain, "/1") {
			wallet.Remain = strings.TrimSuffix(wallet.Remain, "/1")
		}
		if !detail && wallet.AccountDenominator != 0 && float64(wallet.AccountNumerator)/float64(wallet.AccountDenominator) != float64(wallet.AccountNumerator/wallet.AccountDenominator) {
			wallet.Remain = fmt.Sprintf("%0.2f", float64(wallet.AccountNumerator)/float64(wallet.AccountDenominator)/float64(multiInt))
		}
		if !detail {
			if wallet.AccountDenominator == 0 {
				continue
			}
			f := float64(wallet.AccountNumerator) / float64(wallet.AccountDenominator)
			f = math.Abs(f)
			if f < 0.000000001 {
				continue
			}
		}
		body = append(body, []string{wallet.Name, wallet.Remain})
	}
	return album.ToAlbum([]string{"NAME", "MONEY-" + CurrencyName[currencyType]}, body)
}

func PrivateList(db *xorm.Engine, userId int64, detail bool, page int) (a telebot.Album, err error) {
	timezone := 8
	{
		ok := false
		ok, err = db.SQL("select timezone from user where id=?", userId).Get(&timezone)
		if err != nil {
			return
		}
		if !ok {
			err = fmt.Errorf("user not found")
			return
		}
	}
	type Wallet struct {
		Type              int
		BeforeNumerator   int64
		BeforeDenominator int64
		Inc               float64
		AfterNumerator    int64
		AfterDenominator  int64
		Command           string
		CreateTime        time.Time
	}
	limit := 50
	offset := (page - 1) * limit
	logs := make([]*Wallet, 0)
	err = db.SQL(""+
		"select"+
		" wallet_log.id"+
		",wallet_log.type"+
		",wallet_log.before_numerator"+
		",wallet_log.before_denominator"+
		",wallet_log.inc"+
		",wallet_log.after_numerator"+
		",wallet_log.after_denominator"+
		",command.command"+
		",wallet_log.create_time"+
		" from wallet_log"+
		" left join command on command.id=wallet_log.command_id"+
		" where user_id=?"+
		" order by create_time desc"+
		" limit ? offset ?", userId, limit, offset).Find(&logs)
	if err != nil {
		return
	}
	pageTitle := ""
	{
		ok := false
		count := 0
		ok, err = db.SQL(""+
			"select"+
			" count(*) from wallet_log"+
			" where user_id=?", userId).Get(&count)
		if err != nil {
			return
		}
		if !ok {
			return
		}
		pageCount := int(math.Ceil(float64(count) / float64(limit)))
		pageTitle = fmt.Sprintf("%d/%d", page, pageCount)
	}
	body := make([][]string, 0, 5)
	for _, log := range logs {
		before := big.NewRat(log.BeforeNumerator, log.BeforeDenominator)
		after := big.NewRat(log.AfterNumerator, log.AfterDenominator)
		inc := new(big.Rat).Sub(after, before)

		f, _ := before.Float64()
		beforeStr := MarshalCurrencyNumber(int64(f), log.Type)
		if before.Denom().Cmp(big.NewInt(1)) == 0 {
			beforeStr = MarshalCurrencyNumber(before.Num().Int64(), log.Type)
		} else if detail {
			beforeStr = MarshalCurrencyNumber(before.Num().Int64(), log.Type) + "/" + before.Denom().String()
		}

		f, _ = after.Float64()
		afterStr := MarshalCurrencyNumber(int64(f), log.Type)
		if after.Denom().Cmp(big.NewInt(1)) == 0 {
			afterStr = MarshalCurrencyNumber(after.Num().Int64(), log.Type)
		} else if detail {
			afterStr = MarshalCurrencyNumber(after.Num().Int64(), log.Type) + "/" + after.Denom().String()
		}

		f, _ = inc.Float64()
		incStr := MarshalCurrencyNumber(int64(f), log.Type)
		if inc.Denom().Cmp(big.NewInt(1)) == 0 {
			incStr = MarshalCurrencyNumber(inc.Num().Int64(), log.Type)
		} else if detail {
			incStr = MarshalCurrencyNumber(inc.Num().Int64(), log.Type) + "/" + inc.Denom().String()
		}
		// 如果日志时间小于于2023年3月25号的话，需要固定减去7小时
		if log.CreateTime.Before(time.Date(2023, 3, 25, 0, 0, 0, 0, time.UTC)) {
			log.CreateTime = log.CreateTime.Add(time.Hour * time.Duration(-7))
		}
		// 根据用户设定的时区，进行时间转换
		log.CreateTime = log.CreateTime.Add(time.Duration(timezone) * time.Hour)
		body = append(body, []string{log.CreateTime.Format("06-01-02 15:04:05") + " " + CurrencyName[log.Type],
			beforeStr,
			incStr,
			afterStr,
			log.Command})
	}
	return album.ToAlbum([]string{pageTitle + " DATE", "BEFORE", "INC", "AFTER", "COMMAND"}, body)
}
