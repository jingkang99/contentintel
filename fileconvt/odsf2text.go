package fileconvt

import (
	"io"

	"github.com/jingkang99/contentintel/fileconvt/ods"
)

func ConvertOds(file io.ReaderAt, size int64, writer io.Writer, limit int64) (written int64, err error) {

	var doc ods.Doc

	f, err := ods.NewReader(file, size)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	if err := f.ParseContent(&doc); err != nil {
		return 0, err
	}

	for n, sheet := range doc.Table {
		rows := sheet.Strings()
		if err = writeOutput(writer, []byte(xlGenerateSheetTitle(sheet.Name, n, int(len(rows)))), &written, &limit); err != nil || limit == 0 {
			return written, err
		}

		for _, row := range rows {

			rowText := ""

			// go through all columns
			for m, text := range row {
				if text != "" {
					text = cleanCell(text)

					if m > 0 {
						rowText += ", "
					}
					rowText += text
				}
			}

			rowText += "\n"

			if err = writeOutput(writer, []byte(rowText), &written, &limit); err != nil || limit == 0 {
				return written, err
			}
		}
	}

	return written, nil
}

// ODS2Cells converts an ODS file to individual cells
// Size is the full size of the input file.
func ODS2Cells(file io.ReaderAt, size int64) (cells []string, err error) {

	var doc ods.Doc

	f, err := ods.NewReader(file, size)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if err := f.ParseContent(&doc); err != nil {
		return nil, err
	}

	for _, sheet := range doc.Table {
		for _, row := range sheet.Strings() {
			for _, text := range row {
				if text != "" {
					text = cleanCell(text)
					cells = append(cells, text)
				}
			}
		}
	}

	return
}
