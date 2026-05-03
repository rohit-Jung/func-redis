// Package main -
package main

import (
	"errors"
	"io"
	"log"
	"net"
)

func main() {
	conn, err := net.Listen("tcp", ":7379")
	if err != nil {
		return
	}

	for {
		lsnr, err := conn.Accept() // this is a blocking call
		if err != nil {
			return
		}
		defer lsnr.Close()

		for {
			// number of bytes
			buf := make([]byte, 512)
			n, err := lsnr.Read(buf[:])
			if err != nil {
				// check end of file
				if errors.Is(err, io.EOF) {
					break
				}

				break
			}

			log.Print(string(buf[:n]))
		}
	}
}
