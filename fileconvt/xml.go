package fileconvt

import (
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"fmt"
	"bytes"
	"encoding/xml"
)

// ConvertXML converts an XML file to text.
func ConvertXML(r io.Reader) (string, map[string]string, error) {
	meta := make(map[string]string)
	cleanXML, err := Tidy(r, true)
	if err != nil {
		return "", nil, fmt.Errorf("tidy error: %v", err)
	}
	result, err := XMLToText(bytes.NewReader(cleanXML), []string{}, []string{}, true)
	if err != nil {
		return "", nil, fmt.Errorf("error from XMLToText: %v", err)
	}
	return result, meta, nil
}

// XMLToText converts XML to plain text given how to treat elements.
func XMLToText(r io.Reader, breaks []string, skip []string, strict bool) (string, error) {
	var result string

	dec := xml.NewDecoder(io.LimitReader(r, SizeLimit))
	dec.Strict = strict
	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		switch v := t.(type) {
		case xml.CharData:
			result += string(v)
		case xml.StartElement:
			for _, breakElement := range breaks {
				if v.Name.Local == breakElement {
					result += "\n"
				}
			}
			for _, skipElement := range skip {
				if v.Name.Local == skipElement {
					depth := 1
					for {
						t, err := dec.Token()
						if err != nil {
							// An io.EOF here is actually an error.
							return "", err
						}

						switch t.(type) {
						case xml.StartElement:
							depth++
						case xml.EndElement:
							depth--
						}

						if depth == 0 {
							break
						}
					}
				}
			}
		}
	}
	return result, nil
}

// XMLToMap converts XML to a nested string map.
func XMLToMap(r io.Reader) (map[string]string, error) {
	m := make(map[string]string)
	dec := xml.NewDecoder(io.LimitReader(r, SizeLimit))
	var tagName string
	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch v := t.(type) {
		case xml.StartElement:
			tagName = string(v.Name.Local)
		case xml.CharData:
			m[tagName] = string(v)
		}
	}
	return m, nil
}

// Tidy attempts to tidy up XML.
func Tidy(r io.Reader, xmlIn bool) ([]byte, error) {
	f, err := ioutil.TempFile(os.TempDir(), "docconv")
	if err != nil {
		return nil, err
	}
	defer os.Remove(f.Name())
	io.Copy(f, r)

	var output []byte
	if xmlIn {
		output, err = exec.Command("tidy", "-xml", "-numeric", "-asxml", "-quiet", "-utf8", f.Name()).Output()
	} else {
		output, err = exec.Command("tidy", "-numeric", "-asxml", "-quiet", "-utf8", f.Name()).Output()
	}

	if err != nil && err.Error() != "exit status 1" {
		return nil, err
	}
	return output, nil
}
