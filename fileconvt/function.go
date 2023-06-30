package fileconvt

import (
	"io"
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
			fmt.Printf("%13s\t  %s\n", key, value)
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
