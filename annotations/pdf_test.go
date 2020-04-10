package annotations

import (
	"fmt"
	"testing"
)

func test(name string, t *testing.T) {
	zip := fmt.Sprintf("testfiles/%s.zip", name)
	outfile := fmt.Sprintf("/tmp/%s.pdf", name)
	options := PdfGeneratorOptions{AddPageNumbers: true, AllPages: false, AnnotationsOnly: false}
	generator := CreatePdfGenerator(zip, outfile, options)
	err := generator.Generate()

	if err != nil {
		t.Error(err)
	}
}
func TestGenerateA3(t *testing.T) {
	test("a3", t)
}
func TestGenerateA4(t *testing.T) {
	test("a4", t)
}

func TestGenerateA5(t *testing.T) {
	test("a5", t)
}
func TestGenerateLetter(t *testing.T) {
	test("letter", t)
}
func TestGenerateRM(t *testing.T) {
	test("rm", t)
}
