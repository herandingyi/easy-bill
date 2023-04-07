package internal

import (
	"database/sql"
	"easy-bill/internal/models"
	"encoding/json"
	"errors"
	"xorm.io/xorm"
)

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
