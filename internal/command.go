package internal

import (
	"database/sql"
	"easy-bill/internal/models"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"xorm.io/xorm"
)

func Parse(name string) (currencyType int, remain string) {
	name = " " + strings.TrimSpace(name)
	tokens := []string{" u", " U", " usd", " USD", " 刀", " $", " 美元", " 美金"}
	var token string
	currencyType = 2
	for _, v := range tokens {
		if strings.HasSuffix(name, v) {
			currencyType = 1
			token = v
		}
	}
	remain = strings.TrimSuffix(name, token)
	return
}

func InsertCommand(s *xorm.Session, c *models.Command) (id int64, err error) {
	users := make([]models.User, 0, 5)
	err = s.SQL("select id,name from user where status=1").Find(&users)
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
	ret, err = s.Exec("insert into command"+
		"(sender_id"+
		",command"+
		",args"+
		",create_time) values (?, ?, ?, ?)", c.SenderId, c.Command, body, time.Now())
	if err != nil {
		return 0, err
	}
	return ret.LastInsertId()
}

func QueryPreviousCommand(s *xorm.Session, id int) (int64, error) {
	command := &models.Command{}
	_, err := s.SQL("SELECT id FROM command WHERE "+
		"sender_id = ? "+
		"AND `status` = '1' "+
		"AND ( command.command LIKE '/a%' OR command.command LIKE '/p%' )  "+
		"ORDER BY id DESC LIMIT 1", id).Get(command)
	if err != nil {
		return 0, err
	}
	if command.Id == 0 {
		return 0, errors.New("你还没有账单，无法执行回滚操作")
	}
	_, err = s.Exec("update command set"+
		" status = 2"+
		" where id = ?",
		command.Id)
	if err != nil {
		return 0, err
	}
	return command.Id, nil

}
