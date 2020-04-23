package midiiconv

import (
	"reflect"

	"github.com/djimenez/iconv-go"
	"github.com/m13253/midimark"
)

type stringProcessFunc func(string) string

// DefaultStringProcessFunc if no string process function needed, use it.
func DefaultStringProcessFunc(str string) string {
	return str
}

// Iconv convert text encoding.
func Iconv(seq *midimark.Sequence, fromEncodeing, toEncoding string, spf stringProcessFunc) error {
	for _, track := range seq.Tracks {
		for _, event := range track.Events {
			if mei, ok := event.(midimark.MetaEvent); ok {
				me := reflect.ValueOf(mei).Elem()
				mt := me.FieldByNameFunc(func(f string) bool {
					return f == "Text"
				})
				if mt.IsValid() && mt.CanSet() {
					if ns, err := iconv.ConvertString(mt.String(), fromEncodeing, toEncoding); err != nil {
						return err
					} else {
						ns = spf(ns)
						mt.SetString(ns)
					}
				}
			}
		}
	}
	return nil
}
