package genai

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
)

type TesseractOCR struct{}

func (t *TesseractOCR) ExtractWarrantyAndAMC(doc []byte) (string, string, error) {
	if doc == nil || len(doc) == 0 {
		return "", "", nil
	}
	tmpFile, err := ioutil.TempFile("", "ocr-*.png")
	if err != nil {
		return "", "", err
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(doc); err != nil {
		tmpFile.Close()
		return "", "", err
	}
	tmpFile.Close()

	// Call tesseract CLI: tesseract <infile> stdout
	out, err := exec.Command("tesseract", tmpFile.Name(), "stdout").Output()
	if err != nil {
		return "", "", err
	}
	text := string(out)
	// Find first YYYY-MM-DD
	re := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	m := re.FindString(text)
	if m != "" {
		return m, "", nil
	}
	return "", "", nil
}
