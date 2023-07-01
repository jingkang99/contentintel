package fileconvt

import (
	"os/exec"
	"fmt"
	"time"
	"strings"
)

var pdfTimeLayouts = timeLayouts{time.ANSIC, "Mon Jan _2 15:04:05 2006 MST"}

type timeLayouts []string

type MetaResult struct {
	meta map[string]string
	err  error
}

type BodyResult struct {
	body string
	err  error
}

func ConvertPDF(path string) (string, map[string]string, error) {

	meta := make(map[string]string)
	
	metaStr, err := exec.Command("pdfinfo", path).Output()
	if err != nil {
		return "", nil, err
	}

	for _, line := range strings.Split(string(metaStr), "\n") {
		if parts := strings.SplitN(line, ":", 2); len(parts) > 1 {
			meta[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	// Convert parsed meta
	if x, ok := meta["ModDate"]; ok {
		if t, ok := pdfTimeLayouts.Parse(x); ok {
			meta["ModifiedDate"] = fmt.Sprintf("%d", t.Unix())
		}
	}
	if x, ok := meta["CreationDate"]; ok {
		if t, ok := pdfTimeLayouts.Parse(x); ok {
			meta["CreatedDate"] = fmt.Sprintf("%d", t.Unix())
		}
	}

	body, err1 := exec.Command("pdftotext", "-q", "-nopgbrk", "-enc", "UTF-8", "-eol", "unix", path, "-").Output()
	if err1 != nil {
		return "", nil, err1
	}

	return string(body), meta, nil
}

func ConvertPDF_O(path string) (BodyResult, MetaResult, error) {
	metaResult := MetaResult{meta: make(map[string]string)}
	bodyResult := BodyResult{}
	mr := make(chan MetaResult, 1)
	go func() {
		metaStr, err := exec.Command("pdfinfo", path).Output()
		if err != nil {
			metaResult.err = err
			mr <- metaResult
			return
		}

		// Parse meta output
		for _, line := range strings.Split(string(metaStr), "\n") {
			if parts := strings.SplitN(line, ":", 2); len(parts) > 1 {
				metaResult.meta[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}

		// Convert parsed meta
		if x, ok := metaResult.meta["ModDate"]; ok {
			if t, ok := pdfTimeLayouts.Parse(x); ok {
				metaResult.meta["ModifiedDate"] = fmt.Sprintf("%d", t.Unix())
			}
		}
		if x, ok := metaResult.meta["CreationDate"]; ok {
			if t, ok := pdfTimeLayouts.Parse(x); ok {
				metaResult.meta["CreatedDate"] = fmt.Sprintf("%d", t.Unix())
			}
		}

		mr <- metaResult
	}()

	br := make(chan BodyResult, 1)
	go func() {
		body, err := exec.Command("pdftotext", "-q", "-nopgbrk", "-enc", "UTF-8", "-eol", "unix", path, "-").Output()
		if err != nil {
			bodyResult.err = err
		}

		bodyResult.body = string(body)

		br <- bodyResult
	}()

	return <-br, <-mr, nil
}

func (tl timeLayouts) Parse(x string) (time.Time, bool) {
	for _, layout := range tl {
		t, err := time.Parse(layout, x)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
