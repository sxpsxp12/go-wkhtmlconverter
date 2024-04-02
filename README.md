fork from https://github.com/SebastiaanKlippert/go-wkhtmltopdf
Thanks SebastiaanKlippert for providing inspiration

# go-wkhtmlconverter
Golang commandline wrapper for wkhtmltopdf and wkhtmltoimage

go-wkhtmlconverter is a pure Golang wrapper around the wkhtmltopdf and wkhtmltoimage command line utility.


See http://wkhtmltopdf.org/index.html for wkhtmltopdf docs.

# Installation
go get or use a Go dependency manager of your liking.

```
go get -u github.com/sxpsxp12/go-wkhtmlconverter
```

go-wkhtmltopdf finds the path to wkhtmltopdf by
* first looking in the current dir
* looking in the PATH and PATHEXT environment dirs
* using the WKHTMLTOPDF_PATH environment dir

go-wkhtmltoimage finds the path to wkhtmltopdf by
* first looking in the current dir
* looking in the PATH and PATHEXT environment dirs
* using the WKHTMLTOIMAGE_PATH environment dir

**Warning**: Running executables from the current path is no longer possible in Go 1.19, see https://pkg.go.dev/os/exec@master#hdr-Executables_in_the_current_directory

If you need to set your own wkhtmltopdf and wkhtmltoimage path or want to change it during execution, you can call SetPath().

# Usage

## wkhtml2pdf
See testfile wkhtmltopdf_test.go for more complex options, a common use case test is in simplesample_test.go
```go
package wkhtmlconverter

import (
  "fmt"
  "log"
)

func ExampleNewPDFGenerator() {

  // Create new PDF generator
  pdfg, err := NewPDFGenerator()
  if err != nil {
    log.Fatal(err)
  }

  // Set global options
  pdfg.Dpi.Set(300)
  pdfg.Orientation.Set(OrientationLandscape)
  pdfg.Grayscale.Set(true)

  // Create a new input page from an URL
  page := NewPage("https://godoc.org/github.com/sxpsxp12/go-wkhtmlconverter")

  // Set options for this page
  page.FooterRight.Set("[page]")
  page.FooterFontSize.Set(10)
  page.Zoom.Set(0.95)

  // Add to document
  pdfg.AddPage(page)

  // Create PDF document in internal buffer
  err = pdfg.Create()
  if err != nil {
    log.Fatal(err)
  }

  // Write buffer contents to file on disk
  err = pdfg.WriteFile("./simplesample.pdf")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println("Done")
  // Output: Done
}
```
As mentioned before, you can provide one document from stdin, this is done by using a [PageReader](https://godoc.org/github.com/sxpsxp12/go-wkhtmlconverter#PageReader "GoDoc") object as input to AddPage. This is best constructed with  [NewPageReader](https://godoc.org/github.com/sxpsxp12/go-wkhtmlconverter#NewPageReader "GoDoc") and will accept any io.Reader so this can be used with files from disk (os.File) or memory (bytes.Buffer) etc.  
A simple example snippet:
```go
html := "<html>Hi</html>"
pdfgen.AddPage(NewPageReader(strings.NewReader(html)))
```

## wkhtml2image

See testfile wkhtmltoimage_test.go for more complex options

```angular2html
pImg, err := NewImageGenerator()
if err != nil {
    tb.Fatal(err)
}
pImage.Format.Set("png")
pImage.Height.Set(100)
pImage.Width.Set(100)
pImage.Quality.Set(100)

err = pImg.CreateFromFile("testdata/htmlsimple.html")
if err != nil {
    t.Fatal(err)
}
pImg.WriteFile("testdata/file_htmlsimple.png")
```

# Saving to and loading from JSON

The package now has the possibility to save the PDF Generator object as JSON and to create
a new PDF Generator from a JSON file.
All options and pages are saved in JSON, pages added using NewPageReader are read to memory before saving and then saved as Base64 encoded strings
in the JSON file.

This is useful to prepare a PDF file and generate the actual PDF elsewhere, for example on AWS Lambda.
To create PDF Generator on the client, where wkhtmltopdf might not be present, function `NewPDFPreparer` can be used.

Use `NewPDFPreparer` to create a PDF Generator object on the client and `NewPDFGeneratorFromJSON` to reconstruct it on the server.

```go 
// Client code
pdfg := NewPDFPreparer()
htmlfile, err := ioutil.ReadFile("testdata/htmlsimple.html")
if err != nil {
  log.Fatal(err)
}
    
pdfg.AddPage(NewPageReader(bytes.NewReader(htmlfile)))
pdfg.Dpi.Set(600)
    
// The contents of htmlsimple.html are saved as base64 string in the JSON file
jb, err := pdfg.ToJSON()
if err != nil {
  log.Fatal(err)
}
    
// Server code
pdfgFromJSON, err := NewPDFGeneratorFromJSON(bytes.NewReader(jb))
if err != nil {
  log.Fatal(err)
}
    
err = pdfgFromJSON.Create()
if err != nil {
  log.Fatal(err)
}    
```

For an example of running this in AWS Lambda see https://github.com/sxpsxp12/go-wkhtmlconverter-lambda

# Speed 
The speed if pretty much determined by wkhtmltopdf itself, or if you use external source URLs, the time it takes to get and render the source HTML.

The go wrapper time is negligible with around 0.04ms for parsing an above average number of commandline options.

Benchmarks are included.
