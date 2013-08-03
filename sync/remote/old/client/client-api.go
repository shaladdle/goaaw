package client

import (
	"aaw/sync/remote/common"
)

func (cli *client) Get(id string) error {
	msg := dGetMsg{
		id:    id,
		err:   make(chan error),
	}
	cli.msg <- msg

    return <-msg.err
}

func (cli *client) Put(id string) error {
	msg := dPutMsg{
		id:  id,
		err: make(chan error),
	}
	cli.msg <- msg

	return <-msg.err
}

func (cli *client) Delete(id string) error {
	msg := dDelMsg{
		id:  id,
		err: make(chan error),
	}
	cli.msg <- msg

	return <-msg.err
}

func (cli *client) GetIndex() (map[string]common.FileInfo, error) {
	msg := dGetIndexMsg{
		reply: make(chan map[string]common.FileInfo),
		err:   make(chan error),
	}
	cli.msg <- msg

	select {
	case err := <-msg.err:
		return nil, err
	case r := <-msg.reply:
		return r, nil
	}
}
