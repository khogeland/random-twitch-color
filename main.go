package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"
)

// == Color ==

type Color struct {
	r, g, b byte
}

// "#FFFFFF"
func (c Color) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b)
}

// I'd like to avoid dim colors, so at least one byte will be max value
func randomTwitchColor() Color {
	rand.Seed(time.Now().Unix())
	mask := rand.Perm(3)
	var c [3]byte
	c[mask[0]] = 204
	c[mask[1]] = byte(rand.Intn(204))
	c[mask[2]] = byte(rand.Intn(204))
	return Color{c[0], c[1], c[2]}
}

// ===========

func maybeFail(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func writeOrDie(conn net.Conn, toWrite []byte) {
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err := conn.Write(toWrite)
	maybeFail(err)
}

func login(conn net.Conn, username, token string) {
	writeOrDie(conn, []byte(
		"PASS oauth:"+token+"\nNICK "+username+"\n"))
	var output []byte
	buffer := make([]byte, 1024)
	start := time.Now()
	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buffer)
		maybeFail(err) // Fail even on EOF, I'm not done yet!
		output = append(output, buffer[:n]...)
		if bytes.Contains(output, []byte{'3', '7', '6'}) {
			break
		} else {
			if time.Now().Sub(start).Seconds() >= 15 {
				log.Fatal("Timed out waiting to log in")
			}
		}
	}
}

func setColor(conn net.Conn, username string, color Color) {
	writeOrDie(conn, []byte("PRIVMSG #"+username+" :.color "+color.String()+"\n"))
}

func quit(conn net.Conn) {
	writeOrDie(conn, []byte("QUIT\n"))
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 || len(args) > 2 {
		fmt.Fprintln(os.Stderr, "Usage: random-twitch-color <username> <ouath-token>")
		os.Exit(1)
	}

	username, token := args[0], args[1]

	conn, err := net.Dial("tcp", "irc.twitch.tv:6667")
	defer conn.Close()
	maybeFail(err)

	login(conn, username, token)
	setColor(conn, username, randomTwitchColor())
	quit(conn)

}
