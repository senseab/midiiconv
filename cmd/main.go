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
	fi, fo, ef, et         string
	fixCrLf, charsetDetect bool
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
	flag.BoolVar(&charsetDetect, "charset-detect", false, "detect possible character set.")
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

	if charsetDetect {
		if rs := midiiconv.Detect(seq); err != nil {
			log.Println(err)
		} else {
			for i, c := range rs {
				if c.Err != nil {
					log.Printf("Event #%d::`%s` has error: %v", i, c.Text, c.Err)
				}
				for _, n := range c.Results {
					log.Printf("Event #%d::`%s` has possible charset %s(%d%%)\n", i, c.Text, n.Charset, n.Confidence)
				}
			}
		}
		return
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
