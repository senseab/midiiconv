package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/m13253/midimark"
	"github.com/tonychee7000/midiiconv"
)

var (
	fi, fo, ef, et string
	fixCrLf        bool
)

func warningFunc(err error) {
	log.Println(err)
}

func init() {
	flag.StringVar(&fi, "input", "", "input midi file")
	flag.StringVar(&fo, "output", "", "output midi file")
	flag.StringVar(&ef, "from", "utf-8", "from encoding")
	flag.StringVar(&et, "to", "utf-8", "to encoding")
	flag.BoolVar(&fixCrLf, "fix-crlf", false, "convert `\\r` to `\\r\\n`")
	flag.Parse()
	if fi == "" {
		log.Fatalln("no input file specified, use `-input filename`")
	}
}

func main() {
	input, err := os.Open(fi)
	if err != nil {
		log.Fatalln(err)
	}
	defer input.Close()

	seq, err := midimark.DecodeSequenceFromSMF(input, warningFunc)
	if err != nil {
		log.Fatalln(err)
	}

	if err := midiiconv.Iconv(seq, ef, et, func(str string) string {
		if fixCrLf {
			return strings.ReplaceAll(str, "\r", "\r\n")
		}
		return str
	}); err != nil {
		log.Println(err)
	}

	if fo == "" {
		fo = fi
	}

	output, err := os.Create(fo)
	if err != nil {
		log.Fatalln(err)
	}
	defer output.Close()
	seq.EncodeSMF(output)
}
