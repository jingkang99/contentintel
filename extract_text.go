package main

import (
	"os"
	"fmt"
	"bytes"
	"strings"

	"github.com/jingkang99/contentintel/magicbyte"
	"github.com/jingkang99/contentintel/fileconvt"
)

func main() {

	fmt.Println("\n-------- magic byte")
    data := []byte{0xa1, 0xb2, 0xc3, 0xd4, 0x00, 0x00, 0x00, 0x00}

    fileType, err := magicbyte.Lookup(data)
    if err != nil {
        if err == magicbyte.ErrUnknown {
            fmt.Println("File type is unknown")
            os.Exit(1)
        }else{
            panic(err)
        }
    }

    fmt.Printf("File extension:        %s\n", fileType.Extension)
    fmt.Printf("File type description: %s\n", fileType.Description)

	if len(os.Args) < 2 {
		os.Exit(0)
	}
	filename := os.Args[1]

	fh, err := os.Open(filename)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(2)
	}
	defer fh.Close()

	filename = strings.ToLower(filename)

	var header, footer, body string
	var meta map[string]string

	if	strings.HasSuffix(filename, "docx") == true ||	//ppt
		strings.HasSuffix(filename, "dotx") == true ||	//ppt template
		strings.HasSuffix(filename, "docm") == true ||	//ppt with macro
		strings.HasSuffix(filename, "dotm") == true {	//ppt template with macro

		fmt.Println("-------- docx")
		header, footer, body, meta, err = fileconvt.ConvertDocx(fh)

	}else if
		strings.HasSuffix(filename, "pptx") == true ||	//ppt
		strings.HasSuffix(filename, "pptm") == true ||	//ppt macro enabled
		strings.HasSuffix(filename, "ppsx") == true ||	//show
		strings.HasSuffix(filename, "ppsm") == true ||	//show macro enabled
		strings.HasSuffix(filename, "potx") == true ||	//template
		strings.HasSuffix(filename, "potm") == true	{	//template macro enabled
		
		fmt.Println("-------- pptx")
		header, footer = "", ""
		body, meta, err = fileconvt.ConvertPptx(fh)

	}else if
		strings.HasSuffix(filename, "xlsx") == true ||	//exel
		strings.HasSuffix(filename, "xlsm") == true ||	//macro enabled
		strings.HasSuffix(filename, "xlsb") == true ||	//binary 	
		strings.HasSuffix(filename, "xltx") == true ||	//template
		strings.HasSuffix(filename, "xltm") == true	{	//template macro enabled

		fmt.Println("-------- xlsx")

		// header has header and footer info
		// footer has drawing-text-box info
		header, footer, body, meta, err = fileconvt.GetXlsxText(fh)

		// get all cell content
		const SizeLimit = 20 * 1024 * 1024
		stat, _ := fh.Stat()
		buffer := bytes.NewBuffer(make([]byte, 0, SizeLimit))
		fileconvt.ConvertXlsx(fh, stat.Size(), buffer, SizeLimit, -1)
		body = buffer.String()

	}else if strings.HasSuffix(filename, "rtf") == true {
		fmt.Println("-------- rtf")
		
		header, footer = "", ""

		/* tags and table not handled
			buf := new(bytes.Buffer)
			buf.ReadFrom(fh)
			rtf := buf.String()
			body = fileconvt.RTF2Text(rtf)
		*/

		body, meta, err = fileconvt.ConvertRTF(filename)

	}else if strings.HasSuffix(filename, "pdf") == true {
		fmt.Println("-------- pdf")
		
		header, footer = "", ""

		body, meta, _ = fileconvt.ConvertPDF(filename)

	}else if strings.HasSuffix(filename, "doc") || 
			 strings.HasSuffix(filename, "dot") == true {
		fmt.Println("-------- doc")
		
		header, footer = "", ""

		body, meta, err = fileconvt.ConvertDoc(fh)

	}else if strings.HasSuffix(filename, "xls") || 
			 strings.HasSuffix(filename, "xlt") == true {
		fmt.Println("-------- xls")
		
		header, footer = "", ""

		// get all cell content
		const SizeLimit = 20 * 1024 * 1024
		stat, _ := fh.Stat()
		buffer := bytes.NewBuffer(make([]byte, 0, SizeLimit))
		fileconvt.ConvertXlsO(fh, buffer, stat.Size())
		
		body = buffer.String()
		
		meta, err = fileconvt.GetOffice2k3Meta(fh)

	}else if strings.HasSuffix(filename, "ppt") || 
			 strings.HasSuffix(filename, "pot") == true {
		fmt.Println("-------- ppt")
		
		header, footer = "", ""

		//body, meta, err = fileconvt.ConvertDoc(fh)

		meta, err = fileconvt.GetOffice2k3Meta(fh)

	}else if strings.HasSuffix(filename, "htm") || 
			 strings.HasSuffix(filename, "html")||
			 strings.HasSuffix(filename, "xhtml") == true {
		fmt.Println("-------- html")

		header, footer = "", ""

		body, err = fileconvt.ConvertHtml(fh)
	}

	fileconvt.PrintFileText(header, footer, body, meta)
	
	if err != nil {
		fmt.Println(err)
	}
}
