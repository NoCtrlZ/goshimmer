package mockclientlib

import (
	"fmt"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	"go.nanomsg.org/mangos/v3/transport/all"
	"time"
)

type InMessage struct {
	Data []byte
	Uri  string
}

func readSub(uri string, chIn chan InMessage) error {
	var sock mangos.Socket
	var err error

	if sock, err = sub.NewSocket(); err != nil {
		return err
	}
	if err = sock.Dial(uri); err != nil {
		return err
	}
	err = sock.SetOption(mangos.OptionSubscribe, []byte(""))
	if err != nil {
		return err
	}
	fmt.Printf("connected to %s\n", uri)
	var msg []byte
	for {
		msg, err = sock.Recv()
		if err != nil {
			return err
		}
		chIn <- InMessage{
			Data: msg,
			Uri:  uri,
		}
	}
}

func ReadSub(uri string, chIn chan InMessage) {
	for {
		err := readSub(uri, chIn)
		fmt.Printf("disconnected %s: %v\n", uri, err)
		time.Sleep(1 * time.Second)
		fmt.Printf("reconnecting %s\n", uri)
	}
}

func RunPub(port int, chIn chan []byte) error {
	var sock mangos.Socket
	var err error
	if sock, err = pub.NewSocket(); err != nil {
		return fmt.Errorf("can't get new pub socket: %s", err)
	}
	all.AddTransports(sock)

	if err = sock.Listen(fmt.Sprintf("tcp://:%d", port)); err != nil {
		return fmt.Errorf("can't listen on pub socket: %v", err)
	}
	go func() {
		for msg := range chIn {
			if err := sock.Send(msg); err != nil {
				fmt.Printf("can't send to pub: %v\n", err)
				time.Sleep(1 * time.Second)
			}
		}
	}()
	return nil
}
