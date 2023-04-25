package command

import (
	"database/sql"
	"easy-bill/internal/models"
	"easy-bill/internal/service/currency"
	"encoding/json"
	"math/big"
	"regexp"
	"strings"
	"time"
	"xorm.io/xorm"
)

var nameRegexp = regexp.MustCompile(`[a-z]+`)
var SplitRegexp = regexp.MustCompile(`[,，。:;>]`)

func FixTime(t time.Time, timezone int) time.Time {
	t = t.UTC()
	// 如果日志时间小于于2023年3月25号的话，需要固定减去7小时
	if t.Before(time.Date(2023, 3, 25, 0, 0, 0, 0, time.UTC)) {
		t = t.Add(time.Hour * time.Duration(-7))
	}
	// 根据用户设定的时区，进行时间转换
	t = t.Add(time.Duration(timezone) * time.Hour)
	return t
}

func InsertCommand(s *xorm.Session, c *models.Command) (id int64, err error) {
	type MyUser struct {
		Id        int64
		Name      string
		GroupName string
	}
	users := make([]MyUser, 0, 5)
	err = s.SQL("select `id`,`name`,`group_name` from user").Find(&users)
	if err != nil {
		return 0, err
	}
	menus := make([]models.Menu, 0, 5)
	err = s.SQL("select * from menu").Find(&menus)
	if err != nil {
		return 0, err
	}
	m := make(map[string]interface{})
	m["user"] = users
	m["menu"] = menus
	var body []byte
	body, err = json.Marshal(m)
	if err != nil {
		return 0, err
	}
	var ret sql.Result
	ret, err = s.Exec("insert into"+
		" command"+
		"(sender_id"+
		",command"+
		",args"+
		",create_time) values (?, ?, ?, ?)", c.SenderId, c.Command, body, c.CreateTime)
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

type AaCmd struct {
	CurrencyType           int
	UseDefaultCurrencyType bool
	TotalMoney             int64
	Name2PutMoney          map[string]*big.Rat
	Name2UseMoney          map[string]*big.Rat
	Names                  []string
}

func ToAaCmd(text string) (cmd *AaCmd, err error) {
	if strings.HasPrefix(text, "/a") {
		text = text[2:]
	}
	command := strings.ToLower(strings.TrimSpace(text))
	useDefaultCurrencyType := false
	totalMoney := int64(0)
	name2PutMoney := make(map[string]*big.Rat)
	name2UseMoney := make(map[string]*big.Rat)
	names := make([]string, 0, 5)
	var currencyType int
	{
		currencyType, command, useDefaultCurrencyType = currency.Parse(command)
		//  hcx7500, hr, dsj100
		userStr := command
		ua := SplitRegexp.Split(userStr, -1)
		for _, v := range ua {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			//v  :  hcx7500
			ns := nameRegexp.FindAllString(v, -1)
			n := ""
			for _, name := range ns {
				name = strings.TrimSpace(name)
				if name == "" {
					continue
				}
				n = name
				break
			}
			if n == "" {
				continue
			}
			names = append(names, n)
			as := nameRegexp.Split(v, -1)
			a := int64(0)
			for _, account := range as {
				account = strings.TrimSpace(account)
				if account == "" {
					continue
				}
				a, err = currency.UnmarshalCurrencyNumber(account, currencyType)
				if err != nil {
					return
				}
				break
			}
			totalMoney += a
			if name2PutMoney[n] == nil {
				name2PutMoney[n] = big.NewRat(0, 1)
			}
			name2PutMoney[n] = big.NewRat(0, 1).Add(name2PutMoney[n], big.NewRat(a, 1))
		}
		for _, name := range names {
			if name2UseMoney[name] == nil {
				name2UseMoney[name] = big.NewRat(0, 1)
			}
			name2UseMoney[name] = big.NewRat(0, 1).Add(name2UseMoney[name], big.NewRat(totalMoney, int64(len(names))))
		}
	}
	cmd = &AaCmd{
		CurrencyType:           currencyType,
		UseDefaultCurrencyType: useDefaultCurrencyType,
		TotalMoney:             totalMoney,
		Name2PutMoney:          name2PutMoney,
		Name2UseMoney:          name2UseMoney,
		Names:                  names,
	}
	return
}

func GetUserMap(args string) (name2Id map[string]int64, id2Name map[int64]string, err error) {
	type User struct {
		Id   int64
		Name string
	}
	type MyBody struct {
		User []User
	}
	var myBody MyBody
	err = json.Unmarshal([]byte(args), &myBody)
	if err != nil {
		err = json.Unmarshal([]byte(args), &myBody.User)
		if err != nil {
			return
		}
	}
	name2Id = make(map[string]int64)
	id2Name = make(map[int64]string)
	for _, user := range myBody.User {
		name2Id[user.Name] = user.Id
		id2Name[user.Id] = user.Name
	}
	return
}
