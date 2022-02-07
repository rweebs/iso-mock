package main

import (
	"bufio"
	"fmt"
	"github.com/ayvan/iso8583"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"strings"
	"time"
)

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	socket net.Conn
	data   chan []byte
}

var (
	rc00 = "81288888800"
	rc10 = "81288888810"
	rc11 = "81288888811"
	rc12 = "81288888812"
	rc16 = "81288888816"
	rc17 = "81288888817"
	rc18 = "81288888818"
	rc21 = "81288888821"
	rc25 = "81288888825"
	rc26 = "81288888826"
	rc31 = "81288888831"
	rc32 = "81288888832"
	rc33 = "81288888833"
	rc61 = "81288888861"
)

func (manager *ClientManager) start() {
	for {
		select {
		case connection := <-manager.register:
			manager.clients[connection] = true
			fmt.Println("Added new connection!")
		case connection := <-manager.unregister:
			if _, ok := manager.clients[connection]; ok {
				close(connection.data)
				delete(manager.clients, connection)
				fmt.Println("A connection has terminated!")
			}
		case message := <-manager.broadcast:
			for connection := range manager.clients {
				select {
				case connection.data <- message:
				default:
					close(connection.data)
					delete(manager.clients, connection)
				}
			}
		}
	}
}

func (manager *ClientManager) receive(client *Client) {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
			manager.broadcast <- message
		}
	}
}

func (client *Client) receive() {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			client.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func (manager *ClientManager) send(client *Client) {
	defer func() {
		err := client.socket.Close()
		if err != nil {
			log.Println("error while closing socket client")
		}
	}()

	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			//koollog.Info(fmt.Sprintf("xl-pulsa-iso request message : %v", string(message)))

			m := strings.Replace(string(message), " ", "", -1)
			msgSplite := m[1:18]
			var response []byte

			// signon
			if strings.ToLower(msgSplite) == strings.ToLower("qiso0150000170800") {
				response = []byte(` FISO015000017081082200000020000000400000000000000012117475957895700001\n`)
			}

			// payment
			msgSplitePayment := m[2:18]
			if strings.ToLower(msgSplitePayment) == strings.ToLower("ISO0150000170200") {
				//koollog.Info(fmt.Sprintf("check payment ISO0150000170200 : %v", msgSplitePayment))

				if strings.Contains(m, rc00) {
					//koollog.Info(fmt.Sprintf("RC for success case : %v", rc00))
					response = generateISO("00", rc00)
				}
				if strings.Contains(m, rc10) {
					response = generateISO("10", rc10)
				}
				if strings.Contains(m, rc11) {
					response = generateISO("11", rc11)
				}
				if strings.Contains(m, rc12) {
					response = generateISO("12", rc12)
				}
				if strings.Contains(m, rc16) {
					response = generateISO("16", rc16)
				}
				if strings.Contains(m, rc17) {
					response = generateISO("17", rc17)
				}
				if strings.Contains(m, rc18) {
					response = generateISO("18", rc18)
				}
				if strings.Contains(m, rc21) {
					response = generateISO("21", rc21)
				}
				if strings.Contains(m, rc25) {
					response = generateISO("25", rc25)
				}
				if strings.Contains(m, rc26) {
					response = generateISO("26", rc26)
				}
				if strings.Contains(m, rc31) {
					response = generateISO("31", rc31)
				}
				if strings.Contains(m, rc32) {
					response = generateISO("32", rc32)
				}
				if strings.Contains(m, rc33) {
					response = generateISO("33", rc33)
				}
				if strings.Contains(m, rc61) {
					response = generateISO("61", rc61)
				}
			}
			fmt.Println(response)
			_, err := client.socket.Write(response)
			if err != nil {
				log.Println("error while closing socket client")
			}
		}
	}
}

func StartServerMode() {
	fmt.Println("Starting server...")
	listener, error := net.Listen("tcp", "localhost:8091")
	if error != nil {
		fmt.Println(error)
	}
	manager := ClientManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go manager.start()
	for {
		connection, _ := listener.Accept()
		if error != nil {
			fmt.Println(error)
		}
		client := &Client{socket: connection, data: make(chan []byte)}
		manager.register <- client
		timeoutDuration := 3000000 * time.Second
		err := connection.SetReadDeadline(time.Now().Add(timeoutDuration))
		if err != nil {
			fmt.Println("error set read deadline connection")
		}
		go manager.receive(client)
		err = connection.SetWriteDeadline(time.Now().Add(timeoutDuration))
		if err != nil {
			fmt.Println("error set write deadline connection")
		}
		go manager.send(client)
	}
}

func startClientMode() {
	fmt.Println("Starting client...")
	connection, error := net.Dial("tcp", "localhost:12345")
	if error != nil {
		fmt.Println(error)
	}
	client := &Client{socket: connection}
	go client.receive()
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		connection.Write([]byte(strings.TrimRight(message, "\n")))
	}
}

func generateISO(rc string, phoneNumber string) []byte {
	payloadISO := struct {
		PrimaryAccountNumber       *iso8583.Llvar   `field:"2" length:"16"`
		ProcessingCode             *iso8583.Numeric `field:"3" length:"6"`
		TransactionAmount          *iso8583.Numeric `field:"4" length:"12"`
		TransmissionDateAndTime    *iso8583.Numeric `field:"7" length:"10"`
		SystemTraceAuditNumber     *iso8583.Numeric `field:"11" length:"6"`
		LocalTransactionTime       *iso8583.Numeric `field:"12" length:"6"`
		LocalTransactionDate       *iso8583.Numeric `field:"13" length:"4"`
		SettlementDate             *iso8583.Numeric `field:"15" length:"4"`
		MerchantType               *iso8583.Numeric `field:"18" length:"4"`
		AcquiringInstitutionIDCode *iso8583.Llvar   `field:"32" length:"4"`
		RetrievalReferenceNumber   *iso8583.Numeric `field:"37" length:"12"`
		CardAcceptorTerminalID     *iso8583.Numeric `field:"41" length:"16"`
		AdditionalData             *iso8583.Lllvar  `field:"48" length:"120"`
		TransactionCurrencyCode    *iso8583.Numeric `field:"49" length:"3"`
		BankIDCode                 *iso8583.Lllvar  `field:"63" length:"2"`
		ResponseCode               *iso8583.Numeric `field:"39" length:"2"`
	}{
		PrimaryAccountNumber:       iso8583.NewLlvar([]byte("0000000000000000")),
		ProcessingCode:             iso8583.NewNumeric("570000"),
		TransactionAmount:          iso8583.NewNumeric("021600000000"),
		TransmissionDateAndTime:    iso8583.NewNumeric("0826180731"),
		SystemTraceAuditNumber:     iso8583.NewNumeric("000514"),
		LocalTransactionTime:       iso8583.NewNumeric("180731"),
		LocalTransactionDate:       iso8583.NewNumeric("0826"),
		SettlementDate:             iso8583.NewNumeric("0826"),
		MerchantType:               iso8583.NewNumeric("7011"),
		AcquiringInstitutionIDCode: iso8583.NewLlvar([]byte("9750")),
		RetrievalReferenceNumber:   iso8583.NewNumeric("GE2DGMBYGQ4T"),
		CardAcceptorTerminalID:     iso8583.NewNumeric("7874279955272435"),
		AdditionalData:             iso8583.NewLllvar([]byte(fmt.Sprintf("62%s 0002500000B20210826170851360", phoneNumber))),
		TransactionCurrencyCode:    iso8583.NewNumeric("360"),
		BankIDCode:                 iso8583.NewLllvar([]byte("00")),
		ResponseCode:               iso8583.NewNumeric(rc),
	}

	message := iso8583.NewMessageExtended("0210", iso8583.ASCII, false, true, payloadISO)
	message.MtiEncode = iso8583.ASCII
	byteMsg, _ := message.Bytes()

	isoStrMessage := string(byteMsg)
	isoWithXLHeaderAndFooter := fmt.Sprintf("%s%s%s", "ISO015000017", isoStrMessage, "ISO015000017")
	isoMessageWithLength := ISO8583GenerateFullMessageWithLength(isoWithXLHeaderAndFooter)
	byteMsg = []byte(isoMessageWithLength)

	fmt.Println(fmt.Sprintf("raw_request_iso: %v isoStrMessage: %v", isoMessageWithLength, isoStrMessage))

	fmt.Println(fmt.Sprintf("response iso: %v", byteMsg))
	return []byte(byteMsg)
}

// ISO8583GenerateFullMessageWithLength - Generate ISO 8583 Full Message With Header
func ISO8583GenerateFullMessageWithLength(message string) string {
	messageLength := ISO8583EncodeHeaderLength(message)
	return fmt.Sprintf("%s%s", messageLength, message)
}

func ISO8583EncodeHeaderLength(message string) string {
	var (
		asciiFirst  string
		asciiSecond string
	)

	messageLength := len(message)

	if messageLength >= 256 {
		modMsgLengthWith256 := messageLength % 256
		firstChar := (messageLength - modMsgLengthWith256) / 256
		asciiFirst = string(firstChar)
		asciiSecond = string(modMsgLengthWith256)
	} else {
		asciiFirst = string(0)
		asciiSecond = string(messageLength)
	}

	good := NormalizeCharacters(asciiSecond)

	return fmt.Sprintf("%s%s", asciiFirst, string(good))
}

// NormalizeCharacters - normalize character
func NormalizeCharacters(msg string) []byte {
	var good []byte
	for _, c := range msg {
		good = append(good, byte(c))
	}
	return good
}
