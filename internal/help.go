package internal

import (
	"fmt"
	"sort"
)

var I18n = map[string]map[string]string{
	"zh": {
		"zh":              "中文",
		"en":              "英文",
		"ru":              "俄文",
		"DefaultLanguage": "默认语言",
		"DefaultCurrency": "默认货币",
		"HelpDetail": "-----个人的指令帮助----\n\n" +
			"/l 个人账单明细; 例子1: /l | 例子2: /l 2 (尾部的2代表第二页)\n" +
			"/d 分数形式个人账单明细; 例子1: /d 2\n" +
			"/lr 个人账单报告; 例子1: /lr\n" +
			"/dr 分数形式个人账单报告; 例子1: /dr\n" +
			"/timezone 设置个人时区; 例子1: /timezone 8 (尾部的8代表+8时区)\n" +
			"/group_name 昵称和群名对应表\n" +
			"/help 帮助; 例子1:/help | 例子2: /help en (尾部的en代表查看英文帮助文档)\n\n" +
			"-----群中的指令帮助----\n\n" +
			"/join NAME 加入EASY-BILL; 例子1: /join tk\n" +
			"/a AA账单; 例子1: /a tk5100,rzt | 例子2: /a tk5.1,rzt u (尾部的u代表美元)\n" +
			"/p 向某人支付; 例子1: /p tk5.1,rzt u | 例子2: /p tk5/3,rzt u (支持分数)\n" +
			"/l 账单; 例子1: /l | 例子2: /l u \n" +
			"/d 分数形式账单; 例子1: /d | 例子2: /d u\n" +
			"/group_name 昵称和群名对应表\n" +
			"/help 帮助; 例子1:/help | 例子2: /help en\n\n",
		"SupportCurrencyTable": "支持货币表:\n",
		"SupportLanguageTable": "支持语言表:\n",
	},
	"en": {
		"zh":              "Chinese",
		"en":              "English",
		"ru":              "Russian",
		"DefaultLanguage": "Default Language",
		"DefaultCurrency": "Default Currency",
		"HelpDetail": "-----Personal command help----\n\n" +
			"/l Personal bill details; Example 1: /l | Example 2: /l 2 (the 2 at the end represents the second page)\n" +
			"/d Fractional personal bill details; Example 1: /d 2\n" +
			"/lr Personal bill report; Example 1: /lr\n" +
			"/dr Fractional personal bill report; Example 1: /dr\n" +
			"/timezone Set personal time zone; Example 1: /timezone 8 (the 8 at the end represents +8 time zone)\n" +
			"/group_name Nickname and group name correspondence table\n" +
			"/help Help; Example 1:/help | Example 2: /help en (the en at the end represents English help document)\n\n" +
			"-----Group command help----\n\n" +
			"/join NAME Join EASY-BILL; Example 1: /join tk\n" +
			"/a AA bill; Example 1: /a tk5100,rzt | Example 2: /a tk5.1,rzt u (the u at the end represents the dollar)\n" +
			"/p Pay someone; Example 1: /p tk5.1,rzt u | Example 2: /p tk5/3,rzt u (Fractional support)\n" +
			"/l Bill; Example 1: /l | Example 2: /l u \n" +
			"/d Fractional bill; Example 1: /d | Example 2: /d u\n" +
			"/group_name Nickname and group name correspondence table\n" +
			"/help Help; Example 1:/help | Example 2: /help en\n\n",
		"SupportCurrencyTable": "Support currency table:\n",
		"SupportLanguageTable": "Support language table:\n",
	},
	"ru": {
		"zh":              "Китайский",
		"en":              "Английский",
		"ru":              "Русский",
		"DefaultLanguage": "Язык по умолчанию",
		"DefaultCurrency": "Валюта по умолчанию",
		"HelpDetail": "-----Персональная команда помощи----\n\n" +
			"/l Персональная таблица счетов; Пример 1: /l | Пример 2: /l 2 (2 в конце означает вторую страницу)\n" +
			"/d Дробная персональная таблица счетов; Пример 1: /d 2\n" +
			"/lr Персональный отчет о счетах; Пример 1: /lr\n" +
			"/dr Дробный персональный отчет о счетах; Пример 1: /dr\n" +
			"/timezone Установить персональную временную зону; Пример 1: /timezone 8 (8 в конце означает +8 временную зону)\n" +
			"/group_name Таблица соответствия никнеймов и названий групп\n" +
			"/help Помощь; Пример 1:/help | Пример 2: /help en (en в конце означает английский документ помощи)\n\n" +
			"-----Групповая команда помощи----\n\n" +
			"/join NAME Присоединиться к EASY-BILL; Пример 1: /join tk\n" +
			"/a AA счет; Пример 1: /a tk5100,rzt | Пример 2: /a tk5.1,rzt u (u в конце означает доллар)\n" +
			"/p Оплатить кого-то; Пример 1: /p tk5.1,rzt u | Пример 2: /p tk5/3,rzt u (Поддержка дробных)\n" +
			"/l Счет; Пример 1: /l | Пример 2: /l u \n" +
			"/d Дробный счет; Пример 1: /d | Пример 2: /d u\n" +
			"/group_name Таблица соответствия никнеймов и названий групп\n" +
			"/help Помощь; Пример 1:/help | Пример 2: /help en\n\n",
		"SupportCurrencyTable": "Таблица поддерживаемых валют:\n",
		"SupportLanguageTable": "Таблица поддерживаемых языков:\n",
	},
}
var DefaultLanguage = "zh"

func Help(country string) string {
	currentLang := I18n[country]
	if currentLang == nil {
		currentLang = I18n[DefaultLanguage]
	}
	supportToken := currentLang["SupportCurrencyTable"]
	{
		tokenArray := make([]string, 0, len(CurrencyTokens)-1)
		tokenMap := map[string]int{}
		for i, token := range CurrencyTokens {
			if i == 0 {
				continue
			}
			tokenArray = append(tokenArray, token[0])
			tokenMap[token[0]] = i
		}
		sort.Strings(tokenArray)
		for _, token := range tokenArray {
			i := tokenMap[token]
			supportToken += fmt.Sprint(token, ": ", CurrencyName[i])
			if i == DefaultCurrencyType {
				supportToken += "(" + currentLang["DefaultCurrency"] + "); \n"
			} else {
				supportToken += "; \n"
			}
		}
	}
	supportLanguage := currentLang["SupportLanguageTable"]
	{
		var languages []string
		for lang := range I18n {
			languages = append(languages, lang)
		}
		sort.Strings(languages)
		for _, lang := range languages {
			if lang == DefaultLanguage {
				supportLanguage += fmt.Sprint(lang, ": ", currentLang[lang], "(", currentLang["DefaultLanguage"], "); \n")
			} else {
				supportLanguage += fmt.Sprint(lang, ": ", currentLang[lang], "; \n")
			}
		}
	}
	return currentLang["HelpDetail"] +
		"\n" + supportToken +
		"\n" + supportLanguage
}
