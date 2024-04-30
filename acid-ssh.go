package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gliderlabs/ssh"
)

const (
	port       = ":2222"
	maxTimeout = 3 * time.Minute
)

func drawAsciiArt(s ssh.Session, height, width int) {
	asciiArt := fmt.Sprintf(`
   ╱|、           
 (˚ˎ 。7          Hii %s!
  |、˜〵          you should check out my website at https://sipacid.com
 じしˍ,)ノ        
`, s.User())

	lines := strings.Split(asciiArt, "\n")
	topPadding := (height - len(lines)) / 2
	leftPadding := (width-len(lines[2]))/2 - len(lines[2])
	bottomPadding := height - topPadding - len(lines) - 1

	for i := 0; i < topPadding; i++ {
		io.WriteString(s, "\n")
	}

	for _, line := range lines {
		io.WriteString(s, strings.Repeat(" ", leftPadding)+line+"\n")
	}

	for i := 0; i < bottomPadding; i++ {
		io.WriteString(s, "\n")
	}
}

func handleSession(s ssh.Session) {
	pty, _, isAccepted := s.Pty()
	if !isAccepted {
		return
	}

	height := pty.Window.Height
	width := pty.Window.Width

	drawAsciiArt(s, height, width)

	for {
		displayText := "Press any key to exit..."
		io.WriteString(s, strings.Repeat(" ", (width-len(displayText))/2)+displayText+"\n")
		buf := make([]byte, 1)
		_, err := s.Read(buf)
		if err != nil {
			log.Printf("Something went wrong when trying to read from session: %v", err)
			return // Error occurred but I cba
		}

		io.WriteString(s, "\033[2J\033[H")
		drawAsciiArt(s, height, width)

		displayText = "Do you want to exit? [Y/n]: "
		io.WriteString(s, strings.Repeat(" ", (width-len(displayText))/2)+displayText+"\n")
		buf = make([]byte, 1)
		_, err = s.Read(buf)
		if err != nil {
			log.Printf("Something went wrong when trying to read from session: %v", err)
			return // Error occurred but I cba
		}

		if strings.ToLower(string(buf)) == "y" || string(buf) == "\n" {
			return
		}

		io.WriteString(s, "\033[2J\033[H")
		drawAsciiArt(s, height, width)
		continue
	}
}

func getHostKey() string {
	keyPath := "/keys/host.key"
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		newKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Panicf("failed to generate host key: %v", err)
		}

		keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			log.Panicf("failed to save host key: %v", err)
		}
		defer keyOut.Close()

		privateKeyBytes := x509.MarshalPKCS1PrivateKey(newKey)
		pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes})
	}

	return keyPath
}

func main() {
	ssh.Handle(handleSession)

	log.Printf("starting ssh server on port %v...", port)

	server := &ssh.Server{
		Addr:       port,
		Handler:    ssh.DefaultHandler,
		MaxTimeout: maxTimeout,
	}
	server.SetOption(ssh.HostKeyFile(getHostKey()))

	log.Fatal(server.ListenAndServe())
}
