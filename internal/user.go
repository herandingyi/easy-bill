package internal

import (
	"easy-bill/internal/models"
	"fmt"
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

func GetUserInfo(db *xorm.Engine, userId int64) (timezone int, language string, err error) {

	timezone = 8
	language = DefaultLanguage
	{
		user := &models.User{}
		ok := false
		ok, err = db.SQL("select timezone,language from user where id=?", userId).Get(user)
		if err != nil {
			return
		}
		if !ok {
			err = fmt.Errorf("user not found")
			return
		}
		timezone = user.Timezone
		language = user.Language
	}
	return
}
