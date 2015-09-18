package drain

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func SendToDrain(m string, drain string) error {
	// We don't want drain our own log messages. It creates an infinite loop.
	re := regexp.MustCompile("^no-drain")
	match := re.FindStringIndex(m)
	if match != nil {
		return nil
	}

	u, err := url.Parse(drain)
	if err != nil {
		log.Fatal(err)
	}
	uri := u.Host + u.Path
	switch u.Scheme {
	case "syslog":
		sendToSyslogDrain(m, uri)
	case "https":
		sendToHttpsDrain(m, uri)
	default:
		log.Println("no-drain " + u.Scheme + " drain type is not implemented.")
	}
	return nil
}

func sendToHttpsDrain(m string, drain string) error {
	buf := strings.NewReader(m)

	log.Println("no-drain Sending log message to: " + drain)
	resp, err := http.Post(("https://" + drain), "text/plain", buf)
	if err != nil {
		log.Print("no-drain Https Log Error: " + err.Error())
	}
	log.Print("no-drain Https Log Response Status: " + resp.Status)
	resp.Body.Close()
	return nil
}

func sendToSyslogDrain(m string, drain string) error {
	conn, err := net.Dial("udp", drain)
	if err != nil {
		log.Print(err)
	}
	defer conn.Close()
	fmt.Fprintf(conn, m)
	return nil
}

func getopt(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}
