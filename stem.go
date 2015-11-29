package main

import (
	"bufio"
	"bytes"
	"github.com/fjl/go-couchdb"
	"github.com/kljensen/snowball"
	"github.com/kpmy/tier"
	"github.com/kpmy/ypk/halt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

var defOpts tier.Opts = tier.DefaultOpts
var wordsDb *couchdb.DB
var words *Words

type Words struct {
	sync.Mutex
	last int64
	wm   map[string]map[string]int
}

func (w *Words) New() *Words {
	w = new(Words)
	w.wm = make(map[string]map[string]int)
	return w
}

func (w *Words) Load() (err error) {
	w.Lock()
	const docId = "map"
	tmp := make(map[string]interface{})
	if err = wordsDb.Get(docId, &tmp, nil); err == nil {
		for k, _v := range tmp {
			switch v := _v.(type) {
			case map[string]interface{}:
				mm := make(map[string]int)
				for k, _v := range v {
					switch v := _v.(type) {
					case float64:
						mm[k] = int(v)
					default:
						halt.As(100, k, " ", reflect.TypeOf(v))
					}
				}
				w.wm[k] = mm
			case string: //do nothing
			default:
				halt.As(100, k, " ", reflect.TypeOf(v))
			}
		}
	} else if couchdb.NotFound(err) {
		if _, err = wordsDb.Put(docId, tmp, ""); err == nil {
			err = w.Load()
		}
	} else {
		log.Println(err)
	}
	w.Unlock()
	return
}

func (w *Words) Store() {
	w.Lock()
	const docId = "map"
	if rev, err := wordsDb.Rev(docId); err == nil {
		if _, err = wordsDb.Put(docId, w.wm, rev); err != nil {
			log.Println(err)
		}
	}
	w.Unlock()
}

func (w *Words) Sync() {
	go func() {
		last := time.Now().Unix()
		for {
			if last != w.last {
				w.Store()
				last = w.last
			}
			<-time.After(time.Second)
		}
	}()
}

const letterSym = ""

func init() {
	defOpts.IdentContains = func() string {
		return letterSym
	}

	defOpts.IdentStarts = func() string {
		return letterSym
	}

	defOpts.Skip = func(r rune) bool {
		return true
	}

	defOpts.NoStrings = true
	words = words.New()
	words.Load()
	words.Sync()
}

type WordItem struct {
	word string
	lang string
}

func detectLang(s string) (lang string) {
	lm := make(map[string]string)
	for _, r := range strings.ToLower(s) {
		if !strings.ContainsRune(letterSym, r) {
			switch {
			case unicode.IsNumber(r):
				lm["mixed"] = "mixed"
			case unicode.In(r, unicode.Cyrillic):
				lm["russian"] = "russian"
			case unicode.In(r, unicode.Latin):
				lm["english"] = "english"
			}
		}
	}
	if len(lm) == 1 {
		for _, l := range lm {
			lang = l
		}
	} else {
		lang = ""
	}
	return
}

func split(s string) (ret chan WordItem) {
	ret = make(chan WordItem, 0)
	go func() {
		sc := tier.NewScanner(bufio.NewReader(bytes.NewBufferString(s)), defOpts)
		for sc.Error() == nil {
			sym := sc.Get()
			if sym.Code == tier.Ident {
				w := WordItem{word: strings.ToLower(sym.Value)}
				w.lang = detectLang(w.word)
				if w.lang != "" {
					ret <- w
				}
			}
		}
		close(ret)
	}()
	return
}

func Stem(msg string) {
	defer func() {
		recover()
	}()

	for w := range split(msg) {
		if stem, err := snowball.Stem(w.word, w.lang, true); err == nil {
			var m map[string]int
			ok := false
			if m, ok = words.wm[stem]; !ok {
				m = make(map[string]int)
				words.wm[stem] = m
			}
			if x, ok := m[w.word]; ok {
				m[w.word] = x + 1
			} else {
				m[w.word] = 1
			}
		}
	}
	words.last = time.Now().Unix()
	//log.Println(words.wm)
}
