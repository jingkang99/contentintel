package fileconvt

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

func ConvertOdt(r io.Reader) (string, map[string]string, error) {
	meta := make(map[string]string)
	var textBody string

	b, err := ioutil.ReadAll(io.LimitReader(r, SizeLimit))
	if err != nil {
		return "", nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", nil, fmt.Errorf("error unzipping data: %v", err)
	}

	for _, f := range zr.File {
		switch f.Name {
		case "meta.xml":
			rc, err := f.Open()
			if err != nil {
				return "", nil, fmt.Errorf("error extracting '%v' from archive: %v", f.Name, err)
			}
			defer rc.Close()

			meta, err = XMLToMap(rc)
			
			if err != nil {
				return "", nil, fmt.Errorf("error parsing '%v': %v", f.Name, err)
			}

			if tmp, ok := meta["creator"]; ok {
				meta["Author"] = tmp
			}
			if tmp, ok := meta["initial-creator"]; ok {
				meta["Author"] = tmp
			}
			
			if tmp, ok := meta["date"]; ok {
				if t, err := time.Parse("2006-01-02T15:04:05Z", tmp); err == nil {
					meta["ModifiedDate"] = fmt.Sprintf("%d", t.Unix())
				}
			}
			if tmp, ok := meta["creation-date"]; ok {
				if t, err := time.Parse("2006-01-02T15:04:05Z", tmp); err == nil {
					meta["CreatedDate"] = fmt.Sprintf("%d", t.Unix())
				}
			}

		case "content.xml":
			rc, err := f.Open()
			if err != nil {
				return "", nil, fmt.Errorf("error extracting '%v' from archive: %v", f.Name, err)
			}
			defer rc.Close()

			textBody, err = XMLToText(rc, []string{"br", "p", "tab"}, []string{}, true)
			if err != nil {
				return "", nil, fmt.Errorf("error parsing '%v': %v", f.Name, err)
			}
		}
	}

	return textBody, meta, nil
}
