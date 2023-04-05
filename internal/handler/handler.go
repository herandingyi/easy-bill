package handler

import "github.com/tucnak/telebot"

type Handler struct {
	Cmd string
	M   *telebot.Message
}

var taskChan = make(chan *Handler, 100)
var cmdToHandler = make(map[string]func(m *telebot.Message))

func Reg(bot *telebot.Bot, cmd string, f func(m *telebot.Message)) {
	cmdToHandler[cmd] = f
	bot.Handle(cmd, func(m *telebot.Message) {
		taskChan <- &Handler{Cmd: cmd, M: m}
	})
}

func Deal() {
	for {
		select {
		case task := <-taskChan:
			handler := cmdToHandler[task.Cmd]
			handler(task.M)
		}
	}
}
