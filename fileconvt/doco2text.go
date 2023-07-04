// forked from https://github.com/EndFirstCorp/doc2txt and extracts text from DOC files

package fileconvt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/richardlehane/mscfb"
	"github.com/richardlehane/msoleps"
)

// ConvertDoc converts an MS Word .doc to text.
func ConvertDoc(r io.Reader) (string, map[string]string, error) {
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

		// Save output to a file
		outputFile, err := ioutil.TempFile(os.TempDir(), "docO")
		if err != nil {
			// TODO: Remove this.
			log.Println("TempFile Out:", err)
			return
		}
		defer os.Remove(outputFile.Name())

		err = exec.Command("wvText", f.Name(), outputFile.Name()).Run()
		if err != nil {
			log.Println("wvText:", err)
		}

		var buf bytes.Buffer
		_, err = buf.ReadFrom(outputFile)
		if err != nil {
			// TODO: Remove this.
			log.Println("wvText:", err)
		}

		bc <- buf.String()
	}()

	// TODO: Should errors in either of the above Goroutines stop things from progressing?
	body := <-bc
	meta := <-mc
	
	meta, err = ConvertGB2312toUTF8(meta)

	return body, meta, nil
}


func GetOffice2k3Meta(r io.Reader) (map[string]string, error) {
	f, err := NewLocalFile(r)
	if err != nil {
		return nil, fmt.Errorf("error creating local file: %v", err)
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

	meta := <-mc

	mutf, er1 := ConvertGB2312toUTF8(meta)

	return mutf, er1
}

