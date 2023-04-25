package kick

import (
	"errors"
	"github.com/tucnak/telebot"
	"strconv"
	"strings"
	"sync"
	"time"
	"xorm.io/xorm"
)

type Vote struct {
	userIdMap  map[int64]bool
	updateTime time.Time
}

var tickUserId2Vote = &sync.Map{}

func init() {
	go func() {
		for {
			// 每隔7天清理一次
			time.Sleep(7 * 24 * time.Hour)
			deleteKey := make([]int64, 0, 10)
			tickUserId2Vote.Range(func(k, v interface{}) bool {
				vote := v.(*Vote)
				if time.Now().Sub(vote.updateTime) > 72*time.Hour {
					userId := k.(int64)
					deleteKey = append(deleteKey, userId)
				}
				return true
			})
			for _, k := range deleteKey {
				tickUserId2Vote.Delete(k)
			}
		}
	}()
}
func Kick(db *xorm.Engine, m *telebot.Message) (msg string, err error) {
	var name string
	if m.Chat.Type != telebot.ChatPrivate {
		return
	}
	name = strings.ToLower(strings.TrimSpace(m.Payload))
	if name == "" {
		err = errors.New("请输入姓名 如 /kick zs")
		return
	}
	var message interface{}
	message, err = db.Transaction(func(s *xorm.Session) (msg interface{}, err error) {
		tickUserId := int64(0)
		exist := false
		exist, err = s.SQL("select id from user where name = ? and status=1", name).Get(&tickUserId)
		if err != nil {
			return
		}
		if !exist {
			return nil, errors.New("你输入的用户名(" + name +
				")有误, 请核对后重新输入")
		}
		groupCount := 0
		_, err = s.SQL("select count(*) from user where status=1").Get(&groupCount)
		if err != nil {
			return
		}
		//开始投票
		var vote *Vote
		{
			senderId := int64(m.Sender.ID)
			ok := false
			var v interface{}
			v, ok = tickUserId2Vote.Load(tickUserId)
			if ok {
				vote = v.(*Vote)
			}
			if vote == nil {
				vote = &Vote{make(map[int64]bool), time.Now()}
				tickUserId2Vote.Store(tickUserId, vote)
			}
			// 如果距离 上次投票超过 72个小时 则清空投票记录
			if time.Now().Sub(vote.updateTime) > 72*time.Hour {
				vote.userIdMap = make(map[int64]bool)
			}
			vote.userIdMap[senderId] = true
			vote.updateTime = time.Now()
		}
		min := 20
		if groupCount*2/3 < min {
			// 如果人数不足20人 则投票人数为 2/3
			min = groupCount * 2 / 3
		}
		if len(vote.userIdMap) >= min {
			_, err = s.Exec("update user set status = 0 where id = ?", tickUserId)
			if err != nil {
				return
			}
			tickUserId2Vote.Delete(tickUserId)
			msg = "投票成功; \n" + name + " 已被被踢出 EASY-BILL"
		} else {
			msg = "投票成功; \n再有 " + strconv.Itoa(min-len(vote.userIdMap)) +
				" 人投票后, " + name +
				"将被踢出 EASY-BILL, 当前票数: " + strconv.Itoa(len(vote.userIdMap)) + "/" + strconv.Itoa(min) +
				" \n72小时内有效"
		}
		return
	})
	if err != nil {
		return
	}
	msg = message.(string)
	return
}
