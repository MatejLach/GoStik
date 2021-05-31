package lostik

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"go.bug.st/serial"
)

var (
	RadioInitErr = errors.New("problem initialising radio")
	RadioTxErr   = errors.New("failed transmitting data")
	RadioRxErr   = errors.New("failed receiving data")
)

type LoStik struct {
	DevicePortName string
	BaudRate       int
	devicePort     serial.Port
}

func New(devicePortName string, baudRate int) (LoStik, error) {
	devicePort, err := serial.Open(devicePortName, &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
	})
	if err != nil {
		return LoStik{}, err
	}

	return LoStik{
		DevicePortName: devicePortName,
		BaudRate:       baudRate,
		devicePort:     devicePort,
	}, nil
}

func (ls LoStik) RadioInit(initCmds ...string) error {
	// wake the stick up
	err := ls.writeCmd("sys get ver")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = ls.readResp(ctx)
	if err != nil {
		return err
	}

	// give it some breathing space before radio init
	time.Sleep(1 * time.Second)

	if len(initCmds) == 0 {
		initCmds = []string{
			"radio get mod",
			"radio get sf",
			"mac pause",
			"radio set pwr 10",
		}
	}

	err = ls.execCmds(initCmds)
	if err != nil {
		return err
	}

	resp, err := ls.readResp(ctx)
	if err != nil {
		return err
	}

	r := strings.Split(resp, "\n")

	if len(r) != 4 {
		return RadioInitErr
	}

	if r[3] != "ok" {
		return RadioInitErr
	}

	return nil
}

func (ls LoStik) Tx(data []byte) error {
	var resp string
	var err error

	go func(frame []byte) {
		err = ls.writeCmd(fmt.Sprintf("radio tx %s", hex.EncodeToString(frame)))
		if err != nil {
			return
		}

		time.Sleep(300 * time.Millisecond)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		resp, err = ls.readResp(ctx)
		if err != nil {
			return
		}

		if !strings.HasPrefix(resp, "ok") {
			err = RadioTxErr
			return
		}

		resp, err = ls.readResp(ctx)
		if err != nil {
			return
		}

		if !strings.HasSuffix(resp, "radio_tx_ok") {
			err = RadioTxErr
			return
		}
	}(data)

	return err
}

func (ls LoStik) Rx() ([]byte, error) {
	err := ls.writeCmd("radio rx 0")
	if err != nil {
		return nil, err
	}

	time.Sleep(300 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := ls.readResp(ctx)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(resp, "ok") {
		resp, err = ls.readResp(ctx)
		if err != nil {
			return nil, err
		}

		// receiving a radio_err after an ok seems to be a HW bug of sorts, so prepare for receiving again
		if strings.HasPrefix(resp, "radio_err") {
			err := ls.writeCmd("radio rx 0")
			if err != nil {
				return nil, err
			}

			_, err = ls.readResp(ctx)
			if err != nil {
				return nil, err
			}

			resp, err = ls.readResp(ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	if !strings.HasPrefix(resp, "radio_rx") {
		return nil, RadioRxErr
	}

	return hex.DecodeString(strings.TrimSpace(strings.TrimPrefix(resp, "radio_rx")))
}

func (ls LoStik) writeCmd(cmd string) error {
	_, err := ls.devicePort.Write([]byte(fmt.Sprintf("%s\r\n", cmd)))
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	return nil
}

func (ls LoStik) readResp(ctx context.Context) (string, error) {
	var sb strings.Builder
	doneC := make(chan struct{})
	errC := make(chan error)

	go func(done chan<- struct{}, errCh chan<- error) {
		for {
			bts := make([]byte, 100)
			n, err := ls.devicePort.Read(bts)
			if err != nil {
				if err == io.EOF {
					done <- struct{}{}
					return
				}
				errCh <- err
			}

			if n <= 3 {
				done <- struct{}{}
				return
			}

			_, _ = sb.Write(bts[:n])

			if strings.Contains(string(bts[:n]), "ok\r") || strings.HasSuffix(string(bts[:n]), "\r\n") {
				done <- struct{}{}
				return
			}
		}
	}(doneC, errC)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-doneC:
			return strings.TrimSuffix(strings.Replace(sb.String(), "\r", "", -1), "\n"), nil
		case err := <-errC:
			return "", err
		}
	}
}

func (ls LoStik) execCmds(cmds []string) error {
	for _, cmd := range cmds {
		if cmd != "" {
			err := ls.writeCmd(cmd)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
