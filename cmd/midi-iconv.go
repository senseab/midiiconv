package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/m13253/midimark"
	"github.com/tonychee7000/midiiconv"
)

const textTemplate = `
{{- $resultsLength := (len .) -}}
Possible charset order:
{{- range $i, $r := . -}}
{{- " " -}}
{{- $r.Charset}}({{$r.Score}})
{{- if ge (add $i 1) $resultsLength -}}
.
{{- else -}}
, 
{{- end -}}
{{- end -}}
`

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
	flag.BoolVar(&charsetDetect, "charset-detect", false, "[EXPERIMENTAL] detect possible character set.")
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
		stats := make(map[string]int)
		funcMap := template.FuncMap{
			"add": func(a, b int) int {
				return a + b
			},
			"minus": func(a, b int) int {
				return a - b
			},
			"quote": strconv.Quote,
		}
		t := template.Must(template.New("").Funcs(funcMap).Parse(textTemplate))

		rs := midiiconv.Detect(seq)
		for _, c := range rs {
			for _, r := range c.Results {
				if _, ok := stats[r.Charset]; ok {
					stats[r.Charset] += r.Confidence
				} else {
					stats[r.Charset] = r.Confidence
				}
			}
		}

		statsResult := make([]*struct {
			Charset string
			Score   int
		}, 0)
		for k, v := range stats {
			statsResult = append(statsResult, &struct {
				Charset string
				Score   int
			}{
				k, v,
			})
		}
		sort.Slice(statsResult, func(i, j int) bool {
			return statsResult[i].Score > statsResult[j].Score
		})
		buf := bytes.NewBufferString("")
		t.Execute(buf, statsResult)
		log.Println(buf.String())
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
