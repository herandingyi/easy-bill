package internal

import (
	"easy-bill/internal/album"
	"easy-bill/internal/models"
	"easy-bill/internal/service/command"
	currency2 "easy-bill/internal/service/currency"
	"fmt"
	"github.com/tucnak/telebot"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"xorm.io/xorm"
)

func PrivateReportList(db *xorm.Engine, senderId int64, detail bool) (a telebot.Album, err error) {
	timezone := 8
	timezone, _, err = GetUserInfo(db, senderId)
	if err != nil {
		return
	}
	cmds := make([]*models.Command, 0, 600)
	{
		now := time.Now().UTC()
		now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		err = db.SQL(`
SELECT
command.command,
command.args,
command.create_time
FROM command FORCE INDEX (idx_create_time)
where create_time > ? and command.status=1`, now.AddDate(0, -1, -now.Day()+2)).Find(&cmds)
		if err != nil {
			return
		}
		for _, cmd := range cmds {
			cmd.CreateTime = command.FixTime(cmd.CreateTime, timezone)
		}
	}

	initBill := func() (bill []*big.Rat) {
		bill = make([]*big.Rat, len(currency2.CurrencyName))
		for i := range bill {
			bill[i] = big.NewRat(0, 1)
		}
		return
	}
	now := time.Now().UTC()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	now = now.Add(time.Hour * time.Duration(timezone))
	thisMonth := now.AddDate(0, 0, -now.Day()+1)
	lastMonth := now.AddDate(0, -1, -now.Day()+1)
	addBill := func(bill []*big.Rat, aaCmd *command.AaCmd, name string, createTime, start, end time.Time) {
		if createTime.After(start) && createTime.Before(end) {
			spendBill := aaCmd.Name2UseMoney[name]
			if spendBill == nil {
				return
			}
			for _, cmdName := range aaCmd.Names {
				if name == cmdName {
					bill[aaCmd.CurrencyType] = big.NewRat(0, 1).Add(bill[aaCmd.CurrencyType], spendBill)
				}
			}
		}
	}
	todayBill := initBill()
	yesterdayBill := initBill()
	sevenDayBill := initBill()
	thisMonthBill := initBill()
	lastMonthBill := initBill()
	for _, cmd := range cmds {
		if strings.HasPrefix(cmd.Command, "/p") {
			continue
		}
		if strings.HasPrefix(cmd.Command, "/rollback") {
			continue
		}
		var id2Name map[int64]string
		_, id2Name, err = command.GetUserMap(cmd.Args)
		name := id2Name[senderId]
		if name == "" {
			continue
		}
		var aaCmd *command.AaCmd
		aaCmd, err = command.ToAaCmd(cmd.Command)
		if err != nil {
			return
		}
		//今日开销
		addBill(todayBill, aaCmd, name, cmd.CreateTime, now, now.Add(time.Hour*24))
		//昨日开销
		addBill(yesterdayBill, aaCmd, name, cmd.CreateTime, now.AddDate(0, 0, -1), now)
		//近七日开销
		//addBill(sevenDayBill, aaCmd, name, cmd.CreateTime, now.AddDate(0, 0, -6), now.Add(time.Hour*24))
		//本月开销
		addBill(thisMonthBill, aaCmd, name, cmd.CreateTime, thisMonth, now.Add(time.Hour*24))
		//上月开销
		addBill(lastMonthBill, aaCmd, name, cmd.CreateTime, lastMonth, thisMonth)
	}
	header := []string{"货币", "今日", "昨日", "本月日均", "上月日均", "本月开销", "上月开销"}
	body := make([][]string, 0, 5)
	currency := currency2.GetSortCurrency()
	//获取上个月总天数
	lastMonthDay := int(math.Ceil(float64(thisMonth.Sub(lastMonth).Hours()) / 24))
	currMonthDay := int(math.Ceil(float64(now.Add(24*time.Hour).Sub(thisMonth).Hours()) / 24))
	for _, i := range currency {
		if todayBill[i].Denom().Int64() == 1 && todayBill[i].Num().Int64() == 0 &&
			yesterdayBill[i].Denom().Int64() == 1 && yesterdayBill[i].Num().Int64() == 0 &&
			sevenDayBill[i].Denom().Int64() == 1 && sevenDayBill[i].Num().Int64() == 0 &&
			thisMonthBill[i].Denom().Int64() == 1 && thisMonthBill[i].Num().Int64() == 0 &&
			lastMonthBill[i].Denom().Int64() == 1 && lastMonthBill[i].Num().Int64() == 0 {
			continue
		}
		row := make([]string, 0, 5)
		row = append(row, currency2.CurrencyName[i])
		row = append(row, BillToString(todayBill[i], i, detail))
		row = append(row, BillToString(yesterdayBill[i], i, detail))
		row = append(row, BillToString(big.NewRat(1, 1).Mul(thisMonthBill[i], big.NewRat(1, int64(currMonthDay))), i, detail))
		row = append(row, BillToString(big.NewRat(1, 1).Mul(lastMonthBill[i], big.NewRat(1, int64(lastMonthDay))), i, detail))
		row = append(row, BillToString(thisMonthBill[i], i, detail))
		row = append(row, BillToString(lastMonthBill[i], i, detail))
		body = append(body, row)
	}

	return album.ToAlbum(header, body)
	/*

		toDesc := func(bill []*big.Rat) string {
			ret := ""
			for _, i := range currency {
				if bill[i].Num().Int64() == 0 {
					continue
				}
				if ret == "" {
					ret += fmt.Sprintf("%s(%s)", BillToString(bill[i], i, detail), CurrencyName[i])
				} else {
					ret += fmt.Sprintf("    ,    %s(%s)", BillToString(bill[i], i, detail), CurrencyName[i])
				}
			}
			return ret
		}
		desc = ""
		desc += I18n[language]["Today"] + ":" + toDesc(todayBill) + "\n"
		desc += I18n[language]["Yesterday"] + ":" + toDesc(yesterdayBill) + "\n"
		desc += I18n[language]["LastSevenDay"] + ":" + toDesc(sevenDayBill) + "\n"
		desc += I18n[language]["ThisMonth"] + ":" + toDesc(thisMonthBill) + "\n"
		desc += I18n[language]["LastMonth"] + ":" + toDesc(lastMonthBill) + "\n"
	*/
}

func GroupNameList(db *xorm.Engine) (a telebot.Album, err error) {
	users := make([]*models.User, 0, 5)
	err = db.SQL("SELECT " +
		" `name`," +
		" `group_name`" +
		" FROM user where status=1").Find(&users)
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
	decimalPlace := fmt.Sprint(currency2.CurrencyDecimalPlace[currencyType])
	multi := fmt.Sprint(int(math.Pow(10, float64(currency2.CurrencyDecimalPlace[currencyType]))))
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
	decimalPlace := fmt.Sprint(currency2.CurrencyDecimalPlace[currencyType])
	multiInt := int(math.Pow(10, float64(currency2.CurrencyDecimalPlace[currencyType])))
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
	sum := big.NewRat(0, 1)
	for _, wallet := range wallets {
		if strings.HasSuffix(wallet.Remain, "/1") {
			wallet.Remain = strings.TrimSuffix(wallet.Remain, "/1")
		}
		sum = sum.Add(sum, big.NewRat(wallet.AccountNumerator, wallet.AccountDenominator))
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
	title := "MONEY-" + currency2.CurrencyName[currencyType]
	if sum.Num().Int64() != 0 {
		title += "(" + sum.String() + ")"
	}
	return album.ToAlbum([]string{"NAME", title}, body)
}

func BillToString(bill *big.Rat, currencyType int, detail bool) string {
	f, _ := bill.Float64()
	beforeStr := currency2.MarshalCurrencyNumber(int64(f), currencyType)
	if bill.Denom().Cmp(big.NewInt(1)) == 0 {
		beforeStr = currency2.MarshalCurrencyNumber(bill.Num().Int64(), currencyType)
	} else if detail {
		beforeStr = currency2.MarshalCurrencyNumber(bill.Num().Int64(), currencyType) + "/" + bill.Denom().String()
	}
	return beforeStr
}

func PrivateList(db *xorm.Engine, userId int64, detail, withId bool, page int) (a telebot.Album, err error) {
	timezone := 8
	timezone, _, _ = GetUserInfo(db, userId)
	type Wallet struct {
		Id                int64
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
		" command.id"+
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
		beforeStr := currency2.MarshalCurrencyNumber(int64(f), log.Type)
		if before.Denom().Cmp(big.NewInt(1)) == 0 {
			beforeStr = currency2.MarshalCurrencyNumber(before.Num().Int64(), log.Type)
		} else if detail {
			beforeStr = currency2.MarshalCurrencyNumber(before.Num().Int64(), log.Type) + "/" + before.Denom().String()
		}

		f, _ = after.Float64()
		afterStr := currency2.MarshalCurrencyNumber(int64(f), log.Type)
		if after.Denom().Cmp(big.NewInt(1)) == 0 {
			afterStr = currency2.MarshalCurrencyNumber(after.Num().Int64(), log.Type)
		} else if detail {
			afterStr = currency2.MarshalCurrencyNumber(after.Num().Int64(), log.Type) + "/" + after.Denom().String()
		}

		f, _ = inc.Float64()
		incStr := currency2.MarshalCurrencyNumber(int64(f), log.Type)
		if inc.Denom().Cmp(big.NewInt(1)) == 0 {
			incStr = currency2.MarshalCurrencyNumber(inc.Num().Int64(), log.Type)
		} else if detail {
			incStr = currency2.MarshalCurrencyNumber(inc.Num().Int64(), log.Type) + "/" + inc.Denom().String()
		}
		log.CreateTime = command.FixTime(log.CreateTime, timezone)
		commandStr := log.Command
		if withId {
			commandStr = fmt.Sprintf("%d:%s", log.Id, log.Command)
		}
		body = append(body, []string{log.CreateTime.Format("06-01-02 15:04:05") + " " + currency2.CurrencyName[log.Type],
			beforeStr,
			incStr,
			afterStr,
			commandStr})
	}
	return album.ToAlbum([]string{pageTitle + " DATE", "BEFORE", "INC", "AFTER", "COMMAND"}, body)
}
