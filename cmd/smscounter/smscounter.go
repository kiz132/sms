// SPDX-License-Identifier: MIT
//
// Copyright © 2019 Kent Gibson <warthog618@gmail.com>.

// This command provides an example of generating output similar to
// that generated by github.com/danxexe/sms-counter.
// This is non-optimal as it encodes the required Submit TPDUs, and calculates
// the output from them rather than just performing the minimal calculations
// required for the output.
// OTOH CPU is cheap so I've not bothered to add suitably optimised methods to
// the library.
// YMMV.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/warthog618/sms"
	"github.com/warthog618/sms/encoding/tpdu"
)

func main() {
	var msg string
	var nli int
	flag.StringVar(&msg, "message", "", "The message to encode")
	flag.IntVar(&nli, "language", 0, "The NLI of a character set to use in addition to the default")
	flag.Usage = usage
	flag.Parse()
	if msg == "" {
		flag.Usage()
		os.Exit(1)
	}

	options := []sms.EncoderOption(nil)
	if nli != 0 {
		options = append(options, sms.WithCharset(nli))
	}
	pdus, err := sms.Encode([]byte(msg), options...)
	if err != nil {
		log.Println(err)
		return
	}
	alpha, _ := pdus[0].Alphabet()
	lastLen := len(pdus[len(pdus)-1].UD) // valid for 7bit as it is unpacked into octets.
	pm := pdus[0].UDBlockSize()
	var encoding string
	switch alpha {
	case tpdu.Alpha7Bit:
		if hasEscapes(pdus) {
			encoding = "7BIT_EX"
		} else {
			encoding = "7BIT"
		}
	case tpdu.Alpha8Bit:
		encoding = "8BIT"
	case tpdu.AlphaUCS2:
		encoding = "UCS-2"
		lastLen /= 2 // UCS-2 code points
		pm /= 2      // UCS-2 code points
	}
	rem := pm - lastLen
	count := len(pdus)
	totalLen := (pm * (count - 1)) + lastLen
	fmt.Printf("encoding: %s\n", encoding)
	fmt.Printf("messages: %d\n", count)
	fmt.Printf("total length: %d\n", totalLen)
	fmt.Printf("last PDU length: %d\n", lastLen)
	fmt.Printf("per_message: %d\n", pm)
	fmt.Printf("remaining: %d\n", rem)
}

func hasEscapes(pdus []tpdu.TPDU) bool {
	for _, pdu := range pdus {
		for _, d := range pdu.UD {
			if d == 0x1b {
				return true
			}
		}
	}
	return false
}

func usage() {
	fmt.Fprintf(os.Stderr, "smscounter determimes the number of SMS-Submit TPDUs "+
		"required to encode a given message.\n"+
		"The message is encoded using the GSM7 default alphabet, or if necessary\n"+
		"an optionally specified character set, or failing those as UCS-2.\n"+
		"If the message is too long for a single PDU then it is split into several.\n\n"+
		"Usage: smscounter -message <message>\n")
	flag.PrintDefaults()
}
