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
