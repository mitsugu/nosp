package main

import (
	"os"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"errors"
	//"log"
	//"fmt"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/nbd-wtf/go-nostr/nip19"
)

const (
	timeLayout = "2006-01-02 15:04:05 MST"
)

type CONTENTS struct {
	Date    string `json:"date"`
	PubKey  string `json:"pubkey"`
	Content string `json:"content"`
}
type NOSTRLOG struct {
	Id       string
	Contents CONTENTS
}

func main() {
	var wb []NOSTRLOG // 受信データ
	if err:=GetHomeTimeline(&wb);err!=nil{
		panic(err)
	}
	buf := FormatTimelineForDisplay(wb) // buf is string

	/*
	Building User Interface
	*/
	app := tview.NewApplication()
	textView := tview.NewTextView()
	inputField := tview.NewInputField()

	/*
	text view settings
	*/
	textView.SetBorder(true)
	textView.SetTitle("  Nostr Log  ")
	textView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case ':':
				app.SetFocus(inputField)
				return nil
			}
		}
		return event
	})
	textView.SetText(buf)

	/*
	inputField settings
	*/
	inputField.SetLabel(" : ")
	inputField.SetTitle("Command Line").SetBorder(true).SetTitleAlign(tview.AlignLeft)
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			var cmd []string
			_ = ParseCommandLine(inputField.GetText(), &cmd)
			switch cmd[0] {
			case "q", "quit", "exit":
				app.Stop()
			case "getHome":
				/*
				var wb []NOSTRLOG
				if err:=GetHomeTimeline(&wb);err!=nil{
					panic(err)
				}
				*/
			}
			inputField.SetText("")
			return nil
		case tcell.KeyEscape:
			app.SetFocus(textView)
			return nil
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(inputField, 3, 0, true)

	if err := app.SetRoot(flex, true).SetFocus(textView).Run(); err != nil {
		panic(err)
	}
}

/*
ParseCommandLine
*/
func ParseCommandLine(cmd string,str *[]string) error {
	switch (cmd) {
	case "q":
		*str = append(*str,cmd)
	case "quit":
		*str = append(*str,cmd)
	case "exit":
		*str = append(*str,cmd)
	default:
		r := strings.Split(cmd," ")
		switch (r[0]) {
		case "getHome":
			for i := range r {
				switch i {
				case 0:
					*str = append(*str,r[i])
				case 1,2:
					_,err :=strconv.ParseInt(r[i],10,32)
					if err!=nil {
						_, err := time.Parse(timeLayout, r[i])
						if err!=nil {
							return errors.New("Illegal option")
						}
					}
					*str = append(*str,r[i])
				default:
					return errors.New("Too many option")
				}
			}
		default:
			return errors.New("Not support command")
		}
	}
	return nil
}

// 

/*
FormatTimelineForDisplay
*/
func FormatTimelineForDisplay(wb []NOSTRLOG)string{
	var l []string
	for i := range wb {
		d, err := strconv.ParseInt(wb[i].Contents.Date, 10, 64)
		if err != nil {
			d = int64(0)
		}
		ut := time.Unix(d, 0)
		ut.Format("2006-01-02 15:04:05 MST")
		npub, err := nip19.EncodePublicKey(wb[i].Contents.PubKey)
		if err != nil {
			panic(err)
		}
		l = append(l, "---\n")
		l = append(l, strconv.Itoa(i) +"\n")
		l = append(l, ut.Format("2006-01-02 15:04:05 MST")+"\n")
		l = append(l, "@"+npub+"\n")
		l = append(l, strings.Replace(wb[i].Contents.Content, "\\n", "\n", -1))
		l = append(l, "\n\n")
	}
	buf := strings.Join(l, "")
	return buf
}

//

/*
getHomeTimeline {{{
*/
func GetHomeTimeline(wb *[]NOSTRLOG)error{
	p := make(map[string]CONTENTS)
	b, err := load("hoge.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(b), &p)
	if err != nil {
		return err
	}

	cnt := 0
	for i := range p {
		tmp := NOSTRLOG{i, p[i]}
		*wb = append(*wb, tmp)
		cnt++
	}
	return nil
}

// }}}

/*
load {{{
*/
func load(fn string) (string, error) {
	d, err := getDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(d, fn)
	tmp, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	rep := regexp.MustCompile("{\n\"")
	str := rep.ReplaceAllString(string(tmp), "{\"")
	rep = regexp.MustCompile("},\n}\n")
	str = rep.ReplaceAllString(str, "}}")
	rep = regexp.MustCompile("},\n\"")
	str = rep.ReplaceAllString(str, "}, \"")
	str = strings.ReplaceAll(str, "\n", "\\n")

	return str, nil
}

// }}}

/*
getDir {{{
*/
func getDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("Not set HOME environment variables")
	}
	home = filepath.Join(home, "Downloads")
	if _, err := os.Stat(home); err != nil {
		if err = os.Mkdir(home, 0700); err != nil {
			return "", err
		}
	}
	return home, nil
}

// }}}
