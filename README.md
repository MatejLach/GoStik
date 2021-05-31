GoStik
=

GoStik is a Go(lang) package to send and receive data, (byte slices), utilizing the **LoStik** [LoRa](https://en.wikipedia.org/wiki/LoRa) USB end node.

[Purchase LoStik in the UK/EU](https://connectedthings.store/gb/lostik-the-open-source-lora-development-tool.html).

[Purchase LoStik in the US](https://ronoth.com/products/lostik).


Status
-

Currently, point-to-point LoRa is implemented, LoRaWAN is WIP.


P2P usage example:
-

```go
package main

import (
	"fmt"
	"log"

	"github.com/MatejLach/GoStik/lostik"
)

func main() {
	// Initialize sender
	stickTx, err := lostik.New("/dev/ttyUSB0", 57600)
	if err != nil {
		log.Fatalln(err)
	}

	err = stickTx.RadioInit()
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize receiver
	stickRx, err := lostik.New("/dev/ttyUSB1", 57600)
	if err != nil {
		log.Fatalln(err)
	}

	err = stickRx.RadioInit()
	if err != nil {
		log.Fatalln(err)
	}

	// Send some data
	// Tx is non-blocking
	err = stickTx.Tx([]byte("golang.org"))
	if err != nil {
		log.Fatalln(err)
	}

	// Receive the data we just sent with another device
	resp, err := stickRx.Rx()
	if err != nil {
		log.Fatal(err)
	}

	if string(resp) != "golang.org" {
		log.Fatalln("unexpected response from receiver")
	} else {
		fmt.Printf("Received: %s", resp)
	}
}
```

Note that the receiving device has to be initialized and ready to receive at the same moment as data is being sent.
It is generally NOT possible to receive 'past' data - that is data that was sent by someone prior to a device being ready to receive data. 


Contributing
-

Bug reports and pull requests are welcome. Do not hesitate to open a PR / file an issue, or a feature request.
