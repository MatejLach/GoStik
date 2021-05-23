package lostik

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"go.bug.st/serial"
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
	if len(initCmds) == 0 {
		initCmds = []string{
			"sys get ver",
		}
	}

	for _, cmd := range initCmds {
		if cmd != "" {
			err := ls.writeCmd(cmd)
			if err != nil {
				return err
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := ls.readResp(ctx)
	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func (ls LoStik) writeCmd(cmd string) error {
	_, err := ls.devicePort.Write([]byte(fmt.Sprintf("%s\r\n", cmd)))
	if err != nil {
		return err
	}

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

			_, _ = sb.Write(bts)

			if strings.HasPrefix(string(bts), "ok\r") || strings.HasPrefix(string(bts), "invalid_param\r") || strings.HasPrefix(string(bts), "err\r") {
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
			return strings.Trim(sb.String(), "\n"), nil
		case err := <-errC:
			return "", err
		}
	}
}
