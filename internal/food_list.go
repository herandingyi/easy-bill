package internal

import (
	"easy-bill/internal/album"
	"encoding/json"
	"errors"
	"github.com/tucnak/telebot"

	"easy-bill/internal/models"
	"regexp"
	"strings"
	"time"
	"xorm.io/xorm"
)

type MyMenu struct {
	Name  string
	Code  string
	Price string
}

var nameRegexp = regexp.MustCompile(`[a-z]+`)
var numberRegexp = regexp.MustCompile(`[0-9]+`)

func GetUserAndFood(command, body string) (user string, food string, err error) {
	//确定user 和 食物 id
	user = ""
	foodId := ""
	{
		//指令必须是 /f 开头的
		if !strings.HasPrefix(command, "/f") {
			err = errors.New("指令格式错误")
			return
		}
		command = strings.TrimPrefix(command, "/f")
		var list []string
		list = append(list, nameRegexp.FindAllString(command, -1)...)
		user = strings.TrimSpace(list[1])

		list = make([]string, 0)
		list = append(list, numberRegexp.FindAllString(command, -1)...)
		for _, v := range list {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			foodId = v
			break
		}
	}
	food = ""
	{
		type MyMenu struct {
			Code string
			Name string
		}
		type MyBody struct {
			Menu []MyMenu
		}
		var myBody MyBody
		err = json.Unmarshal([]byte(body), &myBody)
		if err != nil {
			return
		}
		codeToName := make(map[string]string)
		for _, v := range myBody.Menu {
			codeToName[v.Code] = v.Name
		}
		food = codeToName[foodId]
	}
	return
}

func FoodList(db *xorm.Engine) (telebot.Album, error) {
	//最近三个小时内的数据
	from := time.Now().Add(-3 * time.Hour).Format("2006-01-02 15:04:05")
	to := time.Now().Format("2006-01-02 15:04:05")
	commands := make([]*models.Command, 0)
	err := db.SQL("" +
		"select" +
		" command,args FROM command" +
		" where command.`status` = '1' " +
		" AND command.command LIKE '/f %'" +
		" AND command.create_time>='" + from + "'" +
		" and command.create_time<='" + to + "'" +
		"" +
		" ORDER BY id DESC").Find(&commands)
	if err != nil {
		return nil, err
	}
	body := make([][]string, 0, 5)
	//统计
	for _, command := range commands {
		var user, food string
		user, food, err = GetUserAndFood(command.Command, command.Args)
		if err != nil {
			continue
		}
		body = append(body, []string{user, food})
	}
	return album.ToAlbum([]string{"USER", "FOOD"}, body)
}
