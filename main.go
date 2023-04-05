package main

import (
	"easy-bill/internal/handler"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"easy-bill/internal"
	"easy-bill/internal/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tucnak/telebot"
	"xorm.io/xorm"
)

type ChatId struct {
	Id string `json:"id"`
}

func (c *ChatId) Recipient() string {
	return fmt.Sprint(c.Id)
}

func init() {
	fmt.Print(1)
}

var nameRegexp = regexp.MustCompile(`[a-z]+`)
var numberRegexp = regexp.MustCompile(`[0-9]+\.?[0-9]*`)
var splitRegexp = regexp.MustCompile(`[,，。:;>]`)

func main() {
	token := os.Getenv("EASY_BILL_TG_TOKEN")
	dsn := os.Getenv("EASY_BILL_MYSQL_DSN")
	bot, err := telebot.NewBot(telebot.Settings{
		Token: token,
		Poller: &telebot.LongPoller{
			Timeout: 10 * time.Second,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 连接MySQL数据库
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		log.Panic(err)
	}
	handler.Reg(bot, "/join", func(m *telebot.Message) {
		var err error
		var name string
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, name+"成功加入")
			}

		}()
		if m.Chat.Type == telebot.ChatGroup {
			name = strings.ToLower(strings.TrimSpace(m.Payload))
			if name == "" {
				err = errors.New("请输入姓名 如 /join zs")
				return
			}
			if nameRegexp.MatchString(name) {
				if len(name) > 5 {
					err = errors.New("请输入姓名不能超过五个字符")
					return
				}
				_, err = db.Exec("insert into user(id,name,status) values(?,?,1) on duplicate key update status=1", m.Sender.ID, name)
				if err != nil {
					return
				}
			} else {
				err = errors.New("姓名只能包含小写字母")
				return
			}
		}
	})
	// 账单明细 从最近一次结清开始显示
	handler.Reg(bot, "/start", func(m *telebot.Message) {
		if m.Chat.Type == telebot.ChatPrivate {
			log.Printf("[%s] %s", m.Chat.ID, m.Text)
		}
	})
	// 设置时区
	handler.Reg(bot, "/tz", func(m *telebot.Message) {
		var err error
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, "时区设置成功,时区为 "+m.Payload)
			}
		}()
		if m.Chat.Type == telebot.ChatPrivate {
			var timezone int
			if m.Payload == "" {
				timezone = 8
			} else {
				timezone, _ = strconv.Atoi(m.Payload)
			}
			_, err = db.Exec("update user set timezone=? where id=?", timezone, m.Sender.ID)
		}
	})
	//账单
	privateListFunc := func(m *telebot.Message, detail bool) {
		var err error
		if m.Chat.Type != telebot.ChatPrivate {
			return
		}
		page := 1
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				var bill telebot.Album
				bill, err = internal.PrivateList(db, int64(m.Sender.ID), detail, page)
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
				}
			}
		}()
		if m.Payload != "" {
			page, err = strconv.Atoi(m.Payload)
			if err != nil {
				return
			}
		}
	}
	//账单明细 从最近一次结清开始显示
	handler.Reg(bot, "/list", func(m *telebot.Message) { privateListFunc(m, false) })
	handler.Reg(bot, "/list-detail", func(m *telebot.Message) { privateListFunc(m, true) })

	//账单
	listFunc := func(m *telebot.Message, detail bool) {
		var err error
		if m.Chat.Type != telebot.ChatGroup {
			return
		}
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				var bill telebot.Album
				var currencyType int
				currencyType, _ = internal.Parse(strings.TrimSpace(m.Payload))
				bill, err = internal.WalletList(db, currencyType, detail)
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
				}
			}
		}()
	}
	handler.Reg(bot, "/l", func(m *telebot.Message) { listFunc(m, false) })
	handler.Reg(bot, "/detail", func(m *telebot.Message) { listFunc(m, true) })
	//支付/收款
	handler.Reg(bot, "/p", func(m *telebot.Message) {
		var err error
		var currencyType int
		if m.Chat.Type != telebot.ChatGroup {
			return
		}
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				var bill telebot.Album
				bill, err = internal.WalletList(db, currencyType, false)
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
				}
			}
		}()
		command := strings.ToLower(strings.TrimSpace(m.Payload))
		currencyType, command = internal.Parse(command)
		// hcx,7500,dsj
		number := int64(0)
		{
			ns := numberRegexp.FindAllString(command, -1)
			for _, n := range ns {
				n = strings.TrimSpace(n)
				if n == "" {
					continue
				}
				number, err = internal.UnmarshalCurrencyNumber(n, currencyType)
				if err != nil {
					return
				}
			}
			if number == 0 {
				err = errors.New("未探测到数字")
				return
			}
		}
		from := ""
		to := ""
		{
			ns0 := numberRegexp.Split(command, -1)
			ns1 := make([]string, 0, len(ns0))
			for _, u := range ns0 {
				u = strings.TrimSpace(u)
				if u == "" {
					continue
				}
				ns1 = append(ns1, nameRegexp.FindAllString(u, -1)...)
			}
			ns2 := make([]string, 0, len(ns0))
			for _, u := range ns1 {
				u = strings.TrimSpace(u)
				if u == "" {
					continue
				}
				ns2 = append(ns2, u)
			}
			if len(ns2) != 2 {
				err = errors.New("未探测到两个人")
				return
			}
			from = ns2[0]
			to = ns2[1]
		}
		_, err = db.Transaction(func(s *xorm.Session) (_ interface{}, err error) {
			exist := false
			exist, err = s.SQL("select id from user where id=? and name = ? and status=1", m.Sender.ID, from).Exist()
			if err != nil {
				return nil, err
			}
			if !exist {
				return nil, errors.New("付款人必须是发起人")
			}
			toId := int64(0)
			exist, err = s.SQL("select id from user where name = ? and status=1", to).Get(&toId)
			if err != nil {
				return nil, err
			}
			if !exist {
				return nil, errors.New("你输入的用户名(" + to +
					")有误, 请核对后重新输入;\n或是让你伙伴(" + to +
					")输入 /join 加入")
			}
			var commandId int64
			commandId, err = internal.InsertCommand(s, &models.Command{
				SenderId: int64(m.Sender.ID),
				Command:  m.Text})
			if err != nil {
				return nil, err
			}
			err = internal.WalletInc(s, int64(m.Sender.ID), currencyType, number, 1, commandId, m.Time())
			if err != nil {
				return nil, err
			}
			err = internal.WalletInc(s, toId, currencyType, -number, 1, commandId, m.Time())
			if err != nil {
				return nil, err
			}
			return
		})
	})

	//订单
	handler.Reg(bot, "/a", func(m *telebot.Message) {
		var err error
		var currencyType int
		if m.Chat.Type != telebot.ChatGroup {
			return
		}
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				var bill telebot.Album
				bill, err = internal.WalletList(db, currencyType, false)
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
				}
			}
		}()
		command := strings.ToLower(strings.TrimSpace(m.Payload))
		currencyType, command = internal.Parse(command)
		//  hcx7500, hr, dsj100
		userStr := command
		ua := splitRegexp.Split(userStr, -1)
		name2Acount := make(map[string]*big.Rat)
		names := make([]string, 0, 5)
		totalAccount := int64(0)
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
				a, err = internal.UnmarshalCurrencyNumber(account, currencyType)
				if err != nil {
					return
				}
				break
			}
			totalAccount += a
			name2Acount[n] = big.NewRat(a, 1)
		}

		_, err = db.Transaction(func(s *xorm.Session) (_ interface{}, err error) {
			userCount := int64(len(names))
			userIds := make([]int64, 0, userCount)
			userId2Acount := make(map[int64]*big.Rat)
			for _, name := range names {
				var userId int64
				var exist bool
				exist, err = s.SQL("select id from user where name =? and status=1", name).Get(&userId)
				if err != nil {
					return nil, err
				}
				if !exist {
					return nil, errors.New("你输入的用户名(" + name +
						")有误, 请核对后重新输入;\n或是让你伙伴(" + name +
						")输入 /join 加入")
				}
				userIds = append(userIds, userId)
				userId2Acount[userId] = name2Acount[name]
			}
			var commandId int64
			commandId, err = internal.InsertCommand(s, &models.Command{
				SenderId: int64(m.Sender.ID),
				Command:  m.Text})
			if err != nil {
				return nil, err
			}

			for _, userId := range userIds {
				account := userId2Acount[userId]
				gainAccount := new(big.Rat).Sub(account, big.NewRat(totalAccount, userCount))
				err = internal.WalletInc(s, userId, currencyType, gainAccount.Num().Int64(), gainAccount.Denom().Int64(), commandId, m.Time())
				if err != nil {
					return nil, err
				}
			}
			return
		})
	})

	//回滚
	/*
		handler.Reg(bot,"/rollback", func(m *telebot.Message) {
			var err error
			var currencyType int
			if m.Chat.Type != telebot.ChatGroup {
				return
			}
			defer func() {
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					var bill telebot.Album
					bill, err = internal.WalletList(db, currencyType)
					if err != nil {
						bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
					} else {
						bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
					}
				}
			}()
			_, err = db.Transaction(func(s *xorm.Session) (_ interface{}, err error) {
				exist := false
				uid := m.Sender.ID
				var oCid int64
				var nCId int64
				nCId, err = internal.InsertCommand(s, &models.Command{
					SenderId: int64(m.Sender.ID),
					Command:  m.Text})
				if err != nil {
					return nil, err
				}

				exist, err = s.SQL("select * from user where id=? and status=1", uid).Exist()
				if err != nil {
					return nil, err
				}
				if !exist {
					return nil, errors.New("你还没有添加此功能 输入 /join 用户名 加入")
				}
				oCid, err = internal.QueryPreviousCommand(s, uid)
				if err != nil {
					return nil, err
				}
				currencyType, err = internal.WalletRollback(s, uid, oCid, nCId)
				return
			})

		})
	*/
	//菜单的新增
	handler.Reg(bot, "/fma", func(m *telebot.Message) {
		var err error
		var currencyType int
		if m.Chat.Type != telebot.ChatGroup {
			return
		}
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					var bill telebot.Album
					bill, err = internal.MenuList(db)
					if err != nil {
						bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
					} else {
						bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
					}
				}
			}
		}()
		command := strings.TrimSpace(m.Payload)
		var menuStr string
		currencyType, menuStr = internal.Parse(command)
		me := splitRegexp.Split(menuStr, -1)

		_, err = db.Transaction(func(s *xorm.Session) (_ interface{}, err error) {
			if len(me) != 4 {
				return nil, errors.New("请输入正确的菜单信息 比如：/fma 1,卤肉饭,川渝人家,4.5 usd")
			}
			var price int64
			price, err = internal.UnmarshalCurrencyNumber(me[3], currencyType)
			if err != nil {
				return
			}
			err = internal.InsertMenu(s, &models.Menu{
				Code:         me[0],
				Name:         me[1],
				Supplier:     me[2],
				Price:        price,
				CurrencyType: currencyType,
			})
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	})

	//账单
	handler.Reg(bot, "/fm", func(m *telebot.Message) {
		var err error
		if m.Chat.Type != telebot.ChatGroup {
			return
		}
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				var bill telebot.Album
				bill, err = internal.MenuList(db)
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
				}
			}
		}()
	})

	handler.Reg(bot, "/f", func(m *telebot.Message) {
		var err error
		var currencyType int
		if m.Chat.Type != telebot.ChatGroup {
			return
		}
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				var bill telebot.Album
				bill, err = internal.FoodList(db)
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
				}

			}
		}()
		command := strings.ToLower(strings.TrimSpace(m.Payload))
		// hcx,1,dsj
		ns := numberRegexp.FindAllString(command, -1)
		menu := &models.Menu{}
		{
			if len(ns) == 0 {
				err = errors.New("未探测到菜单编码")
				return
			}
			_, err = db.SQL("SELECT * FROM menu WHERE code = ? ", ns[0]).Get(menu)
			if err != nil {
				return
			}
			if menu.Id == 0 {
				err = errors.New("菜品编码:" + ns[0] + "还未入库，请先执行比如: /fma 1,卤肉饭,川渝人家,4.5 usd 将菜品添加到菜单中 ")
				return
			}
			currencyType = menu.CurrencyType
			if err != nil {
				return
			}
		}
		from := ""
		to := ""
		{
			ns0 := numberRegexp.Split(command, -1)
			ns1 := make([]string, 0, len(ns0))
			for _, u := range ns0 {
				u = strings.TrimSpace(u)
				if u == "" {
					continue
				}
				ns1 = append(ns1, nameRegexp.FindAllString(u, -1)...)
			}
			ns2 := make([]string, 0, len(ns0))
			for _, u := range ns1 {
				u = strings.TrimSpace(u)
				if u == "" {
					continue
				}
				ns2 = append(ns2, u)
			}
			if len(ns2) != 2 {
				err = errors.New("未探测到两个人")
				return
			}
			from = ns2[0]
			to = ns2[1]
		}
		_, err = db.Transaction(func(s *xorm.Session) (_ interface{}, err error) {
			exist := false
			fromId := int64(0)
			exist, err = s.SQL("select id from user where name = ? and status=1", from).Get(&fromId)
			if err != nil {
				return nil, err
			}
			if !exist {
				return nil, errors.New("你输入的用户名(" + to +
					")有误, 请核对后重新输入;\n或是让你伙伴(" + to +
					")输入 /join 加入")
			}
			toId := int64(0)
			exist, err = s.SQL("select id from user where name = ? and status=1", to).Get(&toId)
			if err != nil {
				return nil, err
			}
			if !exist {
				return nil, errors.New("你输入的用户名(" + to +
					")有误, 请核对后重新输入;\n或是让你伙伴(" + to +
					")输入 /join 加入")
			}
			var commandId int64
			commandId, err = internal.InsertCommand(s, &models.Command{
				SenderId: int64(m.Sender.ID),
				Command:  m.Text})
			if err != nil {
				return nil, err
			}
			err = internal.WalletInc(s, fromId, currencyType, menu.Price, 1, commandId, m.Time())
			if err != nil {
				return nil, err
			}
			err = internal.WalletInc(s, toId, currencyType, -menu.Price, 1, commandId, m.Time())
			if err != nil {
				return nil, err
			}
			return

		})
	})

	//点菜订单
	handler.Reg(bot, "/fl", func(m *telebot.Message) {
		var err error
		if m.Chat.Type != telebot.ChatGroup {
			return
		}
		defer func() {
			if err != nil {
				bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
			} else {
				var bill telebot.Album
				bill, err = internal.FoodList(db)
				if err != nil {
					bot.Send(&ChatId{fmt.Sprint(m.Chat.ID)}, fmt.Sprint(err))
				} else {
					bot.SendAlbum(&ChatId{fmt.Sprint(m.Chat.ID)}, bill)
				}

			}
		}()
	})
	go func() {
		handler.Deal()
	}()
	bot.Start()
}
