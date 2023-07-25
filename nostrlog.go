package main

import (
	"os"
	"errors"
	//"log"
	//"fmt"
	"path/filepath"
	"io/ioutil"
	"strings"
	"strconv"
	"time"
	"regexp"
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/nbd-wtf/go-nostr/nip19"
)
type CONTENTS struct {
	Date	string	`json:"date"`
	PubKey	string	`json:"pubkey"`
	Content	string	`json:"content"`
}
type NOSTRLOG struct {
	Id			string
	Contents	CONTENTS
}

func main() {
	p := make(map[string]CONTENTS)
	b, err := load("hoge.json")
	if err!=nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(b), &p)
	if err != nil {
		panic(err)
	}
	var wb []NOSTRLOG
	cnt :=0
	for i:= range p {
		tmp:= NOSTRLOG{i,p[i]}
		wb = append(wb,tmp)
		cnt ++
	}

	var l []string
	for i := range wb {
		d,err := strconv.ParseInt(wb[i].Contents.Date,10,64)
		if err!=nil {
			d=int64(0)
		}
		ut := time.Unix(d,0)
		ut.Format("2006-01-02 15:04:05 MST")
		npub, err := nip19.EncodePublicKey(wb[i].Contents.PubKey)
		if err!=nil{
			panic(err)
		}
		l = append(l, "---\n"+strconv.Itoa(i)+"\n"+ut.Format("2006-01-02 15:04:05 MST")+"\n@"+npub+"\n"+wb[i].Contents.Content+"\n\n")
	}
	buf := strings.Join(l,"")

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
			str := inputField.GetText()
			if str=="quit" || str=="exit"{
				app.Stop()
			}
			textView.SetText(textView.GetText(true) + str + "\n")
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
load {{{
*/
func load(fn string) (string,error) {
	d, err := getDir()
	if err != nil {
		return "",err
	}
	path := filepath.Join(d, fn)
	tmp, err := ioutil.ReadFile(path)
	if err != nil {
		return "",err
	}
	rep := regexp.MustCompile(`},\n}`)
	str := rep.ReplaceAllString(string(tmp), "}\n}")
	str = strings.ReplaceAll(str, "\n", "")

	return str, nil
}

//}}}

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

