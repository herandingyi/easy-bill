package album

import (
	"bytes"
	"github.com/fogleman/gg"
	"github.com/olekukonko/tablewriter"
	"github.com/tucnak/telebot"
	"image/color"
	"image/png"
	"os"
	"strings"
)

func ToAlbum(header []string, body [][]string) (telebot.Album, error) {
	text := FormatText(header, body)
	filename := "output.png"
	err := ToImage(text, filename)
	if err != nil {
		return nil, err
	}
	photo := &telebot.Photo{}
	photo.File = telebot.File{FileLocal: filename}
	return []telebot.InputMedia{photo}, nil
}

func FormatText(header []string, body [][]string) string {
	data := make([][]string, 0, 5)
	data = append(data, header)
	data = append(data, body...)
	var buf bytes.Buffer
	// Create a new tablewriter
	table := tablewriter.NewWriter(&buf)

	// Set the table format
	table.SetHeader(data[0])
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_RIGHT)
	// Add the rows to the table
	for _, row := range data[1:] {
		table.Append(row)
	}
	table.Render()

	return buf.String()
}

func ToImage(text, filename string) (err error) {
	fontPath := "./simsun.ttc"
	fontSize := 14
	textColor := color.Black
	backgroundColor := color.White
	texts := strings.Split(text, "\n")
	height := 10*2 + 20*len(texts)
	width := 0
	{
		headerLen := 0
		if len(texts) > 1 {
			headerLen = len(texts[1])
		} else if len(texts) > 0 {
			headerLen = len(texts[0])
		}
		width = 10*2 + 7*headerLen
	}
	// 创建一个新的画布
	dc := gg.NewContext(width, height)

	// 设置画布背景颜色
	dc.SetColor(backgroundColor)
	dc.Clear()

	// 设置字体和字号
	err = dc.LoadFontFace(fontPath, float64(fontSize))
	if err != nil {
		return
	}

	// 绘制文本
	dc.SetColor(textColor)
	for i, t := range texts {
		dc.DrawString(t, 10, 10+20*float64(i+1))
	}
	var file *os.File
	// 将图像保存为 PNG 文件
	file, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	err = png.Encode(file, dc.Image())
	if err != nil {
		return err
	}
	return
}
