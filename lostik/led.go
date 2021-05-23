package lostik

import (
	"time"
)

func (ls LoStik) RedLedOn() error {
	err := ls.writeCmd("sys set pindig GPIO11 1")
	if err != nil {
		return err
	}

	return nil
}

func (ls LoStik) RedLedOff() error {
	err := ls.writeCmd("sys set pindig GPIO11 0")
	if err != nil {
		return err
	}

	return nil
}

func (ls LoStik) BlueLedOn() error {
	err := ls.writeCmd("sys set pindig GPIO10 1")
	if err != nil {
		return err
	}

	return nil
}

func (ls LoStik) BlueLedOff() error {
	err := ls.writeCmd("sys set pindig GPIO10 0")
	if err != nil {
		return err
	}

	return nil
}

func (ls LoStik) ReceivingLedPattern() error {
	time.Sleep(200 * time.Millisecond)

	err := ls.RedLedOn()
	if err != nil {
		return err
	}

	time.Sleep(200 * time.Millisecond)

	err = ls.RedLedOff()
	if err != nil {
		return err
	}

	return nil
}

func (ls LoStik) SendingLedPattern() error {
	time.Sleep(500 * time.Millisecond)

	err := ls.BlueLedOn()
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	err = ls.BlueLedOff()
	if err != nil {
		return err
	}

	return nil
}

func (ls LoStik) SendingReceivingInterleavingLedPattern() error {
	time.Sleep(500 * time.Millisecond)

	err := ls.SendingLedPattern()
	if err != nil {
		return err
	}

	time.Sleep(1275 * time.Millisecond)

	err = ls.ReceivingLedPattern()
	if err != nil {
		return err
	}

	return nil
}
