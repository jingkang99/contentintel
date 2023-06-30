package comm

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	
	"io"
	"io/ioutil"
)

func GetCmdlineArg(long string, short string) string {
	cmdlineArgs := strings.Join(os.Args[1:], " ")

	str := fmt.Sprintf("(%s|%s)", long, short)
	match, _ := regexp.MatchString(str, cmdlineArgs)
	if match {
		for i := 1; i < len(os.Args); {
			if os.Args[i] == long || os.Args[i] == short {
				if i+1 < len(os.Args) {
					return os.Args[i+1]
				}
				i += 2
			}
			i++
		}
	}
	return ""
}

func GetCmdlineBool(long string, short string) bool {
	cmdlineArgs := strings.Join(os.Args[1:], " ")

	str := fmt.Sprintf("(%s|%s)", long, short)
	match, _ := regexp.MatchString(str, cmdlineArgs)
	if match {
		return true
	} else {
		return false
	}
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
