package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/smtp"
	"os"
	"time"
)

var strFrom = flag.String("f", "", "From Address")
var strTo = flag.String("t", "", "To Address")

const RC = "\033[0;31m"
const GC = "\033[0;32m"
const NC = "\033[0m"

var (
	base36 = []byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
		'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
		'U', 'V', 'W', 'X', 'Y', 'Z'}
)

// Encode encodes a number to base36
func Encode(value uint64) string {

	var res [16]byte
	var i int
	for i = len(res) - 1; value != 0; i-- {
		res[i] = base36[value%36]
		value /= 36
	}
	return string(res[i+1:])
}

func buildMessageId(domain string) (messageId string) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	randBytes := make([]byte, 5)
	_, err := r.Read(randBytes)
	if err != nil {
		panic(fmt.Errorf("Failed to generate Message Id"))
	}

	unixTime := uint64(time.Now().UnixNano())
	unixTime36 := Encode(unixTime)
	messageId = fmt.Sprintf("<%v.%X@%v>", unixTime36, randBytes, domain)

	return
}

func buildMail(servers []string) {
	for i := range servers {
		//Open connection to SMTP server
		server := servers[i]
		domain, _, err := net.SplitHostPort(server)
		if err != nil {
			panic(fmt.Errorf("Failed to get MTA domain name. Exiting"))
		}
		messageId := buildMessageId(domain)
		from := fmt.Sprintf("From: %v\r\n", *strFrom)
		to := fmt.Sprintf("To: %v\r\n", *strTo)
		headers := fmt.Sprintf("DKIM-Signature: to-be-removed\r\nSubject: Test email from %v\r\nMessage-Id: %v\r\n", server, messageId)
		body := fmt.Sprintf("\r\nTest email sent from MTA %v, please ignore.\r\n", server)
		msg := []byte(from + to + headers + body)
		fmt.Printf("Sending to %v...", server)
		err = SendMail(server, nil, *strFrom, []string{*strTo}, msg)
		if err != nil {
			fmt.Printf("%v Failed to send to %v%v\nError: %v\n", RC, server, NC, err)
		} else {
			fmt.Printf("%v Successfully sent email to %v!%v\n", GC, server, NC)
		}
	}
	return
}

func SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()

	if a != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(a); err != nil {
				return err
			}
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
func main() {

	flag.Usage = func() {
		fmt.Printf("Usage: gomailvalidate.go -f <from address> -t <to address> <mailserver:port separated by space>\n\n")
	}
	flag.Parse()
	if flag.NFlag() < 2 || len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	mapServers := flag.Args()
	buildMail(mapServers)
}
