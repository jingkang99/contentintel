package fileconvt

import (
	"io"
	"log"
	"fmt"
	"time"
	"strings"
	"os/exec"

	"github.com/richardlehane/mscfb"
	"github.com/richardlehane/msoleps"
)

// ConvertDoc converts an MS Word .doc to text.
func ConvertPpt(r io.Reader) (string, map[string]string, error) {
	f, err := NewLocalFile(r)
	if err != nil {
		return "", nil, fmt.Errorf("error creating local file: %v", err)
	}
	defer f.Done()

	// meta data
	mc := make(chan map[string]string, 1)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				log.Printf("panic when reading doc format: %v", e)
			}
		}()

		meta := make(map[string]string)

		doc, err := mscfb.New(f)
		if err != nil {
			log.Printf("meta: could not read doc: %v", err)
			return
		}

		props := msoleps.New()
		for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
			if msoleps.IsMSOLEPS(entry.Initial) {
				if oerr := props.Reset(doc); oerr != nil {
					log.Printf("meta: could not reset props: %v", oerr)
					break
				}

				for _, prop := range props.Property {
					meta[prop.Name] = prop.String()
				}
			}
		}

		const defaultTimeFormat = "2006-01-02 15:04:05.999999999 -0700 MST"

		// Convert parsed meta
		if tmp, ok := meta["LastSaveTime"]; ok {
			if t, err := time.Parse(defaultTimeFormat, tmp); err == nil {
				meta["ModifiedDate"] = fmt.Sprintf("%d", t.Unix())
			}
		}
		if tmp, ok := meta["CreateTime"]; ok {
			if t, err := time.Parse(defaultTimeFormat, tmp); err == nil {
				meta["CreatedDate"] = fmt.Sprintf("%d", t.Unix())
			}
		}

		mc <- meta
	}()

	// document body
	bc := make(chan string, 1)
	go func() {

		ppttxt, er1 := exec.Command("./ppthtml", f.Name()).Output()
		if er1 != nil {
			log.Println(" ppthtml:", er1)
		}

		pptReader := strings.NewReader(string(ppttxt))

		utftxt, _ := ConvertHtml(pptReader)

		bc <- utftxt
	}()

	body := <-bc
	meta := <-mc
	
	meta, err = ConvertGB2312toUTF8(meta)

	return body, meta, nil
}
