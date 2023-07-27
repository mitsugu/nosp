package main

import (
	"os"
	"encoding/json"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"path/filepath"
	"io/ioutil"
	"errors"
	//"fmt"
	"github.com/gdamore/tcell/v2"
	//"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/rivo/tview"
)

const (
	timeLayout = "2006-01-02 15:04:05 MST"
	secretDir = ".nostk"
	contactsFile  = "contacts.json"
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
type CONTACT struct {
	Url	 string	`json:"url"`
	Name string	`json:"name"`
}


func main() {
	buf := "Execute getHome command"

	/*
		Building User Interface
	*/
	app := tview.NewApplication()
	flex := tview.NewFlex()
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
			cl := inputField.GetText()
			switch cl {
			case "q", "quit", "exit":
				app.Stop()
			case "initEnv":
			default:
				switch strings.Split(cl, " ")[0] {
				case "getHome":
					var wb []NOSTRLOG
					if err := GetHomeTimeline(&wb, cl); err != nil {
						panic(err)
					}
					buf := FormatTimelineForDisplay(wb) // buf is string
					textView.SetText(buf)
				default:
				}
			}
			inputField.SetText("")
			return nil
		case tcell.KeyEscape:
			app.SetFocus(textView)
			return nil
		}
		return event
	})

	/*
		flex box settings
	*/
	flex.SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(inputField, 3, 0, true)

	if err := app.SetRoot(flex, true).SetFocus(textView).Run(); err != nil {
		panic(err)
	}
}

/*
	FormatTimelineForDisplay {{{
*/
func FormatTimelineForDisplay(wb []NOSTRLOG) string {
	p := make(map[string]CONTACT)
	cbuf,err := load(contactsFile)
	if err!=nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(cbuf), &p)
	if err != nil {
		panic(err)
	}

	var l []string
	for i := range wb {
		d, err := strconv.ParseInt(wb[i].Contents.Date, 10, 64)
		if err != nil {
			d = int64(0)
		}
		ut := time.Unix(d, 0)
		ut.Format("2006-01-02 15:04:05 MST")
		npub := p[wb[i].Contents.PubKey].Name
		if err != nil {
			panic(err)
		}
		l = append(l, "---\n")
		l = append(l, strconv.Itoa(i)+"\n")
		l = append(l, ut.Format("2006-01-02 15:04:05 MST")+"\n")
		l = append(l, "@"+npub+"\n")
		l = append(l, strings.Replace(wb[i].Contents.Content, "\\n", "\n", -1))
		l = append(l, "\n\n")
	}
	buf := strings.Join(l, "")
	return buf
}

// }}}

/*
	GetHomeTimeline {{{
*/
func GetHomeTimeline(wb *[]NOSTRLOG, cl string) error {
	strtmp := ""
	if cl == "getHome" {
		cmd := exec.Command("nostk", "dispHome")
		buf, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		strtmp = string(buf)
	} else {
		str := strings.Replace(cl, "getHome ", "", -1)
		//str = "\"" + str + "\""
		cmd := exec.Command("nostk", "dispHome", str)
		buf, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		strtmp = string(buf)
	}
	rep := regexp.MustCompile("{\n\"")
	str := rep.ReplaceAllString(strtmp, "{\"")
	rep = regexp.MustCompile("},\n}\n")
	str = rep.ReplaceAllString(str, "}}")
	rep = regexp.MustCompile("},\n\"")
	str = rep.ReplaceAllString(str, "}, \"")
	str = strings.ReplaceAll(str, "\\.", ".")
	str = strings.ReplaceAll(str, "\t", "\\t")
	str = strings.ReplaceAll(str, "\n", "\\n")

	p := make(map[string]CONTENTS)
	err := json.Unmarshal([]byte(str), &p)
	if err != nil {
		/* for test
		if err, ok := err.(*json.SyntaxError); ok {
		log.Println(string(str[err.Offset-15:err.Offset+15]))
		}
		*/
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
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	r := strings.ReplaceAll(string(b), "\n", "")
	return r, nil
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
	home = filepath.Join(home, secretDir)
	if _, err := os.Stat(home); err != nil {
		if err = os.Mkdir(home, 0700); err != nil {
			return "", err
		}
	}
	return home, nil
}

// }}}

