package internal

import (
	"github.com/tucnak/telebot"
	"strings"
	"xorm.io/xorm"
)

func GetGroupName(m *telebot.Message) string {
	return strings.TrimSpace(m.Sender.LastName + " " + m.Sender.FirstName)
}

func UpdateGroupName(s *xorm.Session, m *telebot.Message) (err error) {
	_, err = s.Exec("update user set group_name = ? where id = ?",
		GetGroupName(m), m.Sender.ID)
	if err != nil {
		return err
	}
	return nil
}
