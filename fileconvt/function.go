package fileconvt

import (
	"os"
	"os/exec"
	"io"
	"io/ioutil"
	"fmt"
	"regexp"
	"strings"
)
// https://www.fileformat.info/info/unicode/char/3000/index.htm
// UTF-8 (hex)	0xE3 0x80 0x80 (e38080)
func GenSection(sep string) string {
	return string('\u3000') + "[[" + sep + "]]" + string('\u3000')
}

// time format https://zetcode.com/golang/datetime-parse/

// https://stackoverflow.com/questions/71979579/how-to-remove-multiple-line-break-n-from-a-string-but-keep-only-one


func PrintFileText(header string, footer string, body string, meta map[string]string) bool {
	var text = ""

	t := strings.TrimSpace(header)
	if (len(t) > 0){
		text += GenSection("header") + header
	}

	t = strings.TrimSpace(footer)
	if (len(t) > 0){
		text += GenSection("footer") + footer
	}

	t = strings.TrimSpace(body)
	if (len(t) > 0){
		text += GenSection("body") + "\n" + body
	}

	// remove long digit string
	re := regexp.MustCompile(`(\d){21,}`)
	text = re.ReplaceAllString(text, "")

	// remove multiple returns
	re = regexp.MustCompile(`(\r\n?|\n){2,}`)
	text = re.ReplaceAllString(text, "$1$1")

	fmt.Println(text)

	if len(meta) > 0 {
		fmt.Println(GenSection("meta"))
	}

	for key, value := range meta {
		v := strings.TrimSpace(value)
		length := len([]rune(v))
		if(length > 0){
			fmt.Printf("%20s\t  %s\n", key, value)
		}
	}
	
	if (len(text) > 0){
		return true
	}
	return false
}

// cleanCell returns a cleaned cell text without new-lines
func cleanCell(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.TrimSpace(text)

	return text
}

func xlGenerateSheetTitle(name string, number, rows int) (title string) {
	if number > 0 {
		title += "\n"
	}

	title += fmt.Sprintf("Sheet \"%s\" (%d rows):\n", name, rows)

	return title
}

func writeOutput(writer io.Writer, output []byte, alreadyWritten *int64, size *int64) (err error) {

	if int64(len(output)) > *size {
		output = output[:*size]
	}

	*size -= int64(len(output))

	writtenOut, err := writer.Write(output)
	*alreadyWritten += int64(writtenOut)

	return err
}

// wraps an *os.File
type LocalFile struct {
	*os.File
	unlink bool
}

// NewLocalFile ensures that there is a file which contains the data provided by r.  If r is
// actually an instance of *os.File then this file is used, otherwise a temporary file is
// created and the data from r copied into it.  Callers must call Done() when
// the LocalFile is no longer needed to ensure all resources are cleaned up.
func NewLocalFile(r io.Reader) (*LocalFile, error) {
	if f, ok := r.(*os.File); ok {
		return &LocalFile{
			File: f,
		}, nil
	}

	f, err := ioutil.TempFile(os.TempDir(), "docconv")
	if err != nil {
		return nil, fmt.Errorf("error creating temporary file: %v", err)
	}
	_, err = io.Copy(f, r)
	if err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, fmt.Errorf("error copying data into temporary file: %v", err)
	}

	return &LocalFile{
		File:   f,
		unlink: true,
	}, nil
}

// Done cleans up all resources.
func (l *LocalFile) Done() {
	l.Close()
	if l.unlink {
		os.Remove(l.Name())
	}
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func ConvertGB2312toUTF8(meta map[string]string) (map[string]string, error) {
	var text string
	for key, value := range meta {
		v := strings.TrimSpace(value)
		length := len([]rune(v))
		if(length > 0){
			text = text + fmt.Sprintf("%20s\t  %s\n$7$7$7", key, value)
		}
	}

	f, err := os.CreateTemp(os.TempDir(), "encconv")
    check(err)
	defer f.Close()

	_, err = f.WriteString(text)
    check(err)

	f.Sync()
	
	txtUTF8, err := exec.Command("iconv", "-f", "gb2312", "-t", "UTF-8", f.Name()).Output()
	
	if err = os.Remove(f.Name()); err != nil {
		fmt.Errorf("error delete temp: %v", err)
    }

	utxt := strings.Split(string(txtUTF8), "\n$7$7$7")

	mutf := make(map[string]string)
	for i := 0; i < len(utxt); i++ {
		if len( strings.TrimSpace(utxt[i]) ) < 2 {continue}
		
		rs := strings.SplitAfter(utxt[i], "\t")
		mutf[ strings.TrimSpace(rs[0]) ] = strings.TrimSpace(rs[1])
	}

	return mutf, err
}
