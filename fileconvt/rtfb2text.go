// https://github.com/kuhumcst/rtfreader
//
// rtfreader -i ../data/1.rtf -t 1


//https://sourceforge.net/projects/gnuwin32/files/unrtf/0.19.3/unrtf-0.19.3-bin.zip/download
// git clone  https://github.com/ropensci/unrtf
// unrtf --nopict -t text ../data/1.rtf
//
//Chinese NOT handled

package fileconvt

import (
	"os"
	"os/exec"
	"io/ioutil"
	"fmt"
	"time"
	"bytes"
	"strings"
)

// ConvertRTF converts RTF files to text.
func ConvertRTF(fileName string) (string, map[string]string, error) {
	meta := make(map[string]string)

	tmpF, err := ioutil.TempFile(os.TempDir(), "rtfz")
	if err != nil {
		return "", meta ,fmt.Errorf("error creating temp file: %v", err)
	}
	defer os.Remove(tmpF.Name())

	_, err = exec.Command("rtfreader", "-i", fileName, "-t", tmpF.Name()).Output()
	if err != nil {
		return "", nil, fmt.Errorf("rtfreader error: %v", err)
	}

	fByte, _ := os.ReadFile(tmpF.Name())
	body := string(fByte)

	// ---------------  parse meta using unrtf	
	cmd := exec.Command("unrtf", "--nopict", "-t", "text", fileName)
	
	var outb, errb bytes.Buffer
    cmd.Stdout = &outb
    cmd.Stderr = &errb

	err = cmd.Run()
	if err != nil {
		//return "", nil, fmt.Errorf("unrtf error: %v", err)
	}
	outstr := outb.String() + errb.String()

	// Step through content looking for meta data and stripping out comments
	for _, line := range strings.Split(outstr, "\n") {

		if strings.Contains(line, "### creaton date: ") {
			p := strings.SplitN(line, ":", 2)	
			meta["created"] = strings.TrimSpace(p[1])

			//t, _ := time.Parse(time.RFC822, "28 Jun 23 12:35")
			
			t, e := time.Parse("02 January 2006 15:04", meta["created"])
			if e == nil {
				meta["CreatedDate"] = fmt.Sprintf("%d", t.Unix())
			}
			continue
		}else if strings.Contains(line, "### revision date: ") {
			p := strings.SplitN(line, ":", 2)	
			meta["modified"] = strings.TrimSpace(p[1])

			t, e := time.Parse("02 January 2006 15:04", meta["modified"] )
			if e == nil {
				meta["ModifiedDate"] = fmt.Sprintf("%d", t.Unix())
			}
			continue
		}else if strings.HasPrefix(line, "AUTHOR: ") {
			p := strings.SplitN(line, ":", 2)	
			meta["creator"] = strings.TrimSpace(p[1])
			continue
		}else if strings.HasPrefix(line, "TITLE: ") {
			p := strings.SplitN(line, ":", 2)	
			meta["title"] = strings.TrimSpace(p[1])
			continue
		}else if ! strings.HasPrefix(line, "### ") || strings.HasPrefix(line, "### please") {
			continue
		}

		if parts := strings.SplitN(line, ":", 2); len(parts) > 1 {
			key := strings.Replace(parts[0], "### ", "", 1)
			meta[strings.TrimSpace(key)] = strings.TrimSpace(parts[1])
		}
	}

	return body, meta, nil
}

// https://github.com/joniles/rtfparserkit
// java 

// https://github.com/axigenmessaging/rtfconverter
// Go, not work
