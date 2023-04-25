package rollback

import (
	"easy-bill/internal/models"
	"easy-bill/internal/service/command"
	"easy-bill/internal/service/wallet"
	"errors"
	"fmt"
	"github.com/tucnak/telebot"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
	"xorm.io/xorm"
)

type UserVote struct {
	UserName string
	UserId   int64
	Vote     bool
}
type Vote struct {
	Users      []*UserVote
	UpdateTime time.Time
}

var rollbackCommandId2Vote = &sync.Map{}

func init() {
	go func() {
		for {
			// 每隔7天清理一次
			time.Sleep(8 * 24 * time.Hour)
			deleteKey := make([]int64, 0, 10)
			rollbackCommandId2Vote.Range(func(k, v interface{}) bool {
				vote := v.(*Vote)
				if time.Now().Sub(vote.UpdateTime) > 1*time.Hour {
					userId := k.(int64)
					deleteKey = append(deleteKey, userId)
				}
				return true
			})
			for _, k := range deleteKey {
				rollbackCommandId2Vote.Delete(k)
			}
		}
	}()
}

const HELP = "请输入需要撤销id, \n如 /rollback 1:/a tk2,rzt u\n可以在个人命令中输入 /l id，查看COMMAND列"

type WalletLog struct {
	UserId   int64
	UserName string

	Type              int
	BeforeNumerator   int64
	BeforeDenominator int64
	AfterNumerator    int64
	AfterDenominator  int64
}

func Rollback(db *xorm.Engine, m *telebot.Message) (msg string, err error) {
	var cmd string
	cmd = strings.ToLower(strings.TrimSpace(m.Payload))
	if cmd == "" {
		err = errors.New(HELP)
		return
	}

	cmds := strings.Split(cmd, ":")
	if len(cmds) != 2 {
		err = errors.New(HELP)
		return
	}
	var commandId int64
	commandId, err = strconv.ParseInt(cmds[0], 10, 64)
	if err != nil {
		return
	}
	commandArg := cmds[1]
	var message interface{}
	message, err = db.Transaction(func(s *xorm.Session) (msg interface{}, err error) {
		walletLogs := make([]*WalletLog, 0, 10)
		{
			err = s.SQL(`
select user.id as user_id,user.name user_name,
wallet_log.type,
wallet_log.before_numerator,
wallet_log.before_denominator,
wallet_log.after_numerator,
wallet_log.after_denominator
from wallet_log FORCE INDEX (command_id_user_id_index)
left join user on user.id=wallet_log.user_id
where wallet_log.command_id = ?
`, commandId).Find(&walletLogs)
			if err != nil {
				return
			}
			if len(walletLogs) == 0 {
				return nil, errors.New("你输入的命令无法找到")
			}
			senderId := int64(m.Sender.ID)
			containSelf := false
			for _, walletLog := range walletLogs {
				if walletLog.UserId == senderId {
					containSelf = true
					break
				}
			}
			if !containSelf {
				return nil, errors.New("你输入的命令与你无关，无法撤销")
			}
		}
		if len(walletLogs) == 1 {
			if m.Chat.Type == telebot.ChatGroup {
				return nil, errors.New("该撤销命令由只有你一人参与，请在个人聊天中输入")
			}
		} else {
			if m.Chat.Type == telebot.ChatPrivate {
				return nil, errors.New("该撤销命令由多个人参与，请在群组聊天中输入")
			}
		}
		mCommand := &models.Command{}
		// 检查命令, 输入是否合法
		{
			exist := false
			exist, err = s.SQL("select * from command where id = ? for update", commandId).Get(mCommand)
			if err != nil {
				return
			}
			if !exist {
				return nil, errors.New("你输入的命令无法找到")
			}
			if mCommand.Status != 1 {
				return nil, errors.New("你输入的命令已经被撤销")
			}
			if mCommand.Command != commandArg {
				return nil, fmt.Errorf("你输入的命令参数错误, 请重新下面指令输入： /rollback %d:%s", commandId, mCommand.Command)
			}
		}
		//开始投票
		var vote *Vote
		{
			senderId := int64(m.Sender.ID)
			ok := false
			var v interface{}
			v, ok = rollbackCommandId2Vote.Load(commandId)
			if ok {
				vote = v.(*Vote)
			}
			if vote == nil {
				vote = &Vote{
					Users:      make([]*UserVote, 0, 10),
					UpdateTime: time.Now(),
				}
				for _, walletLog := range walletLogs {
					vote.Users = append(vote.Users, &UserVote{
						UserName: walletLog.UserName,
						UserId:   walletLog.UserId,
						Vote:     false,
					})
				}
				rollbackCommandId2Vote.Store(commandId, vote)
			}
			// 如果距离 上次投票超过 1个小时 则清空投票记录
			if time.Now().Sub(vote.UpdateTime) > 1*time.Hour {
				for _, userVote := range vote.Users {
					userVote.Vote = false
				}
			}
			for _, userVote := range vote.Users {
				if userVote.UserId == senderId {
					userVote.Vote = true
					break
				}
			}
			vote.UpdateTime = time.Now()
		}
		voteCount := 0
		{
			for _, userVote := range vote.Users {
				if userVote.Vote {
					voteCount++
				}
			}
		}
		// 全部投票通过
		if voteCount == len(vote.Users) {
			var newCommandId int64
			newCommandId, err = command.InsertCommand(s, &models.Command{
				SenderId:   int64(m.Sender.ID),
				Command:    m.Text,
				CreateTime: m.Time(),
			})
			if err != nil {
				return
			}
			for _, walletLog := range walletLogs {
				inc := big.NewRat(0, 1).Sub(big.NewRat(walletLog.BeforeNumerator, walletLog.BeforeDenominator), big.NewRat(walletLog.AfterNumerator, walletLog.AfterDenominator))
				err = wallet.WalletInc(s, walletLog.UserId,
					walletLog.Type, inc.Num().Int64(), inc.Denom().Int64(),
					newCommandId, m.Time())
			}
			if err != nil {
				return
			}
			_, err = s.Exec("update command set status = 2 where id = ?", commandId)
			if err != nil {
				return
			}
			rollbackCommandId2Vote.Delete(commandId)
			if m.Chat.Type == telebot.ChatGroup {
				msg = "投票成功; \n指令 --- " + commandArg + " --- 已被撤销"
			} else {
				msg = "指令 --- " + commandArg + " --- 已被撤销"
			}
		} else {
			leftPeople := ""
			for i := 0; i < len(vote.Users); i++ {
				if !vote.Users[i].Vote {
					if leftPeople == "" {
						leftPeople += vote.Users[i].UserName
					} else {
						leftPeople += "," + vote.Users[i].UserName
					}
				}
			}
			msg = "投票成功; \n再有 " + leftPeople +
				" 投票后 --- " + commandArg +
				" --- 将被撤销, 当前票数: " + strconv.Itoa(voteCount) + "/" + strconv.Itoa(len(vote.Users)) +
				"\n1小时内有效"
		}
		return
	})
	if err != nil {
		return
	}
	msg = message.(string)
	return
}
