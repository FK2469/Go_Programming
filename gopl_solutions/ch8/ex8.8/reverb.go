// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 224.

// Reverb2 is a TCP server that simulates an echo.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func echo(c net.Conn, shout string, delay time.Duration) {
	fmt.Fprintln(c, "\t", strings.ToUpper(shout))
	time.Sleep(delay)
	fmt.Fprintln(c, "\t", shout)
	time.Sleep(delay)
	fmt.Fprintln(c, "\t", strings.ToLower(shout))
}

//!+
func handleConn(c net.Conn) {
	input := bufio.NewScanner(c)
	
	notIdleCh := make(chan bool)
	go countIdleTime(c, notIdleCh)
	
	for input.Scan() {
	        notIdleCh <- true
		go echo(c, input.Text(), 1*time.Second)
	}
	// NOTE: ignoring potential errors from input.Err()
	c.Close()
}

//!-

func countIdleTime(conn net.Conn, notIdleCh <-chan bool){
        ticker := time.NewTicker(time.Second)
        counter := 0
        max := 20
        for{
                select{
                case <- ticker.C:
                        counter++
                        if counter == max{
                                msg := conn.RemoteAddr().String() + "Idle too long. Kicked out."
                                fmt.Println(msg)
                                fmt.Fprintln(conn,msg)
                                ticker.Stop()
                                conn.Close()
                                return
                        }
                case <- notIdleCh:
                        counter = 0
                }
        }
}

func main() {
	l, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(err) // e.g., connection aborted
			continue
		}
		go handleConn(conn)
	}
}
