package fileconvt

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
	"regexp"
	
	"github.com/jingkang99/contentintel/comm"
)

func GetXlsxText(r io.Reader) (string, string, string, map[string]string, error) {
	var size int64

	// if the reader is a file, avoid loading it all into memory
	var ra io.ReaderAt

	if f, ok := r.(interface {
		io.ReaderAt
		Stat() (os.FileInfo, error)
	}); ok {
		si, err := f.Stat()
		if err != nil {
			return "", "", "", nil, err
		}
		size = si.Size()
		ra = f
	} else {
		b, err := ioutil.ReadAll(io.LimitReader(r, comm.Max_Bytes))
		if err != nil {
			return  "", "", "", nil, nil
		}
		size = int64(len(b))
		ra = bytes.NewReader(b)
	}

	zr, err := zip.NewReader(ra, size)
	if err != nil {
		return  "", "", "", nil, fmt.Errorf("error unzipping data: %v", err)
	}

	zipFiles := mapZipFiles(zr.File)

	contentTypeDefinition, err := getContentTypeDefinition(zipFiles["[Content_Types].xml"])
	if err != nil {
		return  "", "", "", nil, err
	}

	meta := make(map[string]string)
	var textHeader, textBody, textFooter string
	for _, override := range contentTypeDefinition.Overrides {
		f := zipFiles[override.PartName]

		// fmt.Printf("%20s\n", override.PartName)  // /word/document.xml

		switch {
		case override.ContentType == "application/vnd.openxmlformats-package.core-properties+xml":
			rc, err := f.Open()
			if err != nil {
				return  "", "", "", nil, fmt.Errorf("error opening '%v' from archive: %v", f.Name, err)
			}
			defer rc.Close()

			meta, err = XMLToMap(rc)
			if err != nil {
				return  "", "", "", nil, fmt.Errorf("error parsing '%v': %v", f.Name, err)
			}

			if tmp, ok := meta["modified"]; ok {
				if t, err := time.Parse(time.RFC3339, tmp); err == nil {
					meta["ModifiedDate"] = fmt.Sprintf("%d", t.Unix())
				}
			}
			if tmp, ok := meta["created"]; ok {
				if t, err := time.Parse(time.RFC3339, tmp); err == nil {
					meta["CreatedDate"] = fmt.Sprintf("%d", t.Unix())
				}
			}
		
		// textBox 
		case override.ContentType == "application/vnd.openxmlformats-officedocument.drawing+xml":

			txtBox, err := parseDocxText(f)
			if err != nil {
				return  "", "", "", nil, err
			}
			textFooter += txtBox + "\n"

		// text in sheet header and footer
		case override.ContentType == "application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml":

			sheet, err := parseDocxText(f)
			if err != nil {
				return  "", "", "", nil, err
			}
			textHeader += sheet + "\n"

		// docx dotx 
		case override.ContentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml" ||
			 override.ContentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.template.main+xml" ||
			 override.ContentType == "application/vnd.ms-word.document.macroEnabled.main+xml" 							||
			 override.ContentType == "application/vnd.ms-word.template.macroEnabledTemplate.main+xml"  :
			
			continue
			
			body, err := parseDocxText(f)
			if err != nil {
				return  "", "", "", nil, err
			}
			textBody += body + "\n"
		case override.ContentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml":
			continue

			footer, err := parseDocxText(f)
			if err != nil {
				return  "", "", "", nil, err
			}
			textFooter += footer + "\n"
		case override.ContentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml":
			continue
			
			header, err := parseDocxText(f)
			if err != nil {
				return  "", "", "", nil, err
			}
			textHeader += header + "\n"
		}
	}
	
	re := regexp.MustCompile(`&\w`)
	
	textHeader = re.ReplaceAllString(textHeader, " ")

	return textHeader, textFooter, textBody, meta, nil
}
