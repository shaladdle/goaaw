package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type EventInfo struct {
	Name    string
	NumArgs int
}

type Notifiable interface {
	Notify(id string, args ...string) error
}

type UI struct {
	not   Notifiable
	fatal chan error
	cmds  map[string]EventInfo
}

func New(not Notifiable, events []EventInfo) (*UI, error) {
	ret := &UI{
		not:   not,
		cmds:  make(map[string]EventInfo),
		fatal: make(chan error),
	}

	for _, info := range events {
		ret.cmds[info.Name] = info
	}

	go ret.repl()

	return ret, nil
}

func (ui *UI) repl() {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Enter a command:")
		b, hasMoreInLine, err := in.ReadLine()
		if hasMoreInLine {
			fmt.Println("hasMoreLine = true")
			continue
		}

		if err != nil {
			fmt.Println("ReadLine error:", err)
			continue
		}

		line := string(b)
		tokens := strings.Split(line, " ")
		fmt.Println("tokens", tokens)
		if len(tokens) < 2 {
			fmt.Println("Invalid command, try again")
			continue
		}
		if info, exists := ui.cmds[tokens[0]]; exists {
			if len(tokens) != 1+info.NumArgs {
				fmt.Printf("Wrong number of arguments for '%v', expected %v, got %v\n",
					tokens[0], info.NumArgs, len(tokens)+1)
				continue
			}
		} else {
			fmt.Println("Command", tokens[0], "not recognized")
			continue
		}

		if err := ui.not.Notify(tokens[0], tokens[1:]...); err != nil {
			fmt.Println("Notify error:", err)
			continue
		}

		fmt.Println("Command executed")
	}
}

func (ui *UI) Fatal() <-chan error {
	return ui.fatal
}
