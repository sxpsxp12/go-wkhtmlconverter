package wkhtmltopdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var imageBinPath stringStore

// SetImagePath sets the path to wkhtmltoimage
func SetImagePath(path string) {
	imageBinPath.Set(path)
}

// GetImagePath gets the path to wkhtmltoimage
func GetImagePath() string {
	return imageBinPath.Get()
}

// ImageGenerator is the main wkhtmltoimage struct, always use NewImageGenerator to obtain a new ImageGenerator struct
type ImageGenerator struct {
	imageOptions

	input      io.Reader
	inputFile  string //filename to read from,Can be a url (http://example.com), a local file (/tmp/example.html)
	OutputFile string //filename to write to, default empty (writes to internal buffer)

	binPath   string
	outbuf    bytes.Buffer
	outWriter io.Writer
	stdErr    io.Writer
}

// Args returns the commandline arguments as a string slice
func (imgg *ImageGenerator) Args() []string {
	imgg.Quiet.Set(true)

	args := append([]string{}, imgg.imageOptions.Args()...)

	if imgg.inputFile != "" {
		args = append(args, imgg.inputFile)
	} else {
		args = append(args, "-")
	}

	if imgg.OutputFile != "" {
		args = append(args, imgg.OutputFile)
	} else {
		args = append(args, "-")
	}
	return args
}

// ArgString returns Args as a single string
func (imgg *ImageGenerator) ArgString() string {
	return strings.Join(imgg.Args(), " ")
}

// Buffer returns the embedded output buffer used if OutputFile is empty
func (imgg *ImageGenerator) Buffer() *bytes.Buffer {
	return &imgg.outbuf
}

// Bytes returns the output byte slice from the output buffer used if OutputFile is empty
func (imgg *ImageGenerator) Bytes() []byte {
	return imgg.outbuf.Bytes()
}

// SetOutput sets the output to write the PDF to, when this method is called, the internal buffer will not be used,
// so the Bytes(), Buffer() and WriteFile() methods will not work.
func (imgg *ImageGenerator) SetOutput(w io.Writer) {
	imgg.outWriter = w
}

// SetStderr sets the output writer for Stderr when running the wkhtmltoimage command. You only need to call this when you
// want to print the output of wkhtmltoimage (like the progress messages in verbose mode). If not called, or if w is nil, the
// output of Stderr is kept in an internal buffer and returned as error message if there was an error when calling wkhtmltoimage.
func (imgg *ImageGenerator) SetStderr(w io.Writer) {
	imgg.stdErr = w
}

// WriteFile writes the contents of the output buffer to a file
func (imgg *ImageGenerator) WriteFile(filename string) error {
	return os.WriteFile(filename, imgg.Bytes(), 0666)
}

// findPath finds the path to wkhtmltoimage by
// - first looking in the current dir
// - looking in the PATH and PATHEXT environment dirs
// - using the WKHTMLTOIMAGE_PATH environment dir
// Warning: Running executables from the current path is no longer possible in Go 1.19
// See https://pkg.go.dev/os/exec@master#hdr-Executables_in_the_current_directory
// The path is cached, meaning you can not change the location of wkhtmltoimage in
// a running program once it has been found
func (imgg *ImageGenerator) findPath() error {
	const exe = "wkhtmltoimage"
	imgg.binPath = GetImagePath()
	if imgg.binPath != "" {
		// wkhtmltoimage has already been found, return
		return nil
	}
	exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	path, err := lookPath(filepath.Join(exeDir, exe))
	if err == nil && path != "" {
		imageBinPath.Set(path)
		imgg.binPath = path
		return nil
	}
	path, err = lookPath(exe)
	if errors.Is(err, exec.ErrDot) {
		return err
	}
	if err == nil && path != "" {
		imageBinPath.Set(path)
		imgg.binPath = path
		return nil
	}
	dir := os.Getenv("WKHTMLTOIMAGE_PATH")
	if dir == "" {
		return fmt.Errorf("%s not found", exe)
	}
	path, err = lookPath(filepath.Join(dir, exe))
	if errors.Is(err, exec.ErrDot) {
		return err
	}
	if err == nil && path != "" {
		imageBinPath.Set(path)
		imgg.binPath = path
		return nil
	}
	return fmt.Errorf("%s not found", exe)
}

func (imgg *ImageGenerator) checkDuplicateFlags() error {
	// we currently can only have duplicates in the global options, so we only check these
	var options []string
	for _, arg := range imgg.imageOptions.Args() {
		if strings.HasPrefix(arg, "--") { // this is not ideal, the value could also have this prefix
			for _, option := range options {
				if option == arg {
					return fmt.Errorf("duplicate argument: %s", arg)
				}
			}
			options = append(options, arg)
		}
	}
	return nil
}

// CreateFromFile creates the image document and stores it in the internal buffer if no error is returned
// filename to read from,Can be a url (http://example.com), a local file (/tmp/example.html)
func (imgg *ImageGenerator) CreateFromFile(filename string) error {
	imgg.inputFile = filename
	imgg.input = nil
	return imgg.run(context.Background())
}

func (imgg *ImageGenerator) CreateFromIOReader(input io.Reader) error {
	imgg.input = input
	imgg.inputFile = ""
	return imgg.run(context.Background())
}

// CreateContext is Create with a context passed to exec.CommandContext when calling wkhtmltoimage
func (imgg *ImageGenerator) CreateContext(ctx context.Context) error {
	return imgg.run(ctx)
}

func (imgg *ImageGenerator) run(ctx context.Context) error {
	// check for duplicate flags
	err := imgg.checkDuplicateFlags()
	if err != nil {
		return err
	}

	// create command
	cmd := exec.CommandContext(ctx, imgg.binPath, imgg.Args()...)

	// set stderr to the provided writer, or create a new buffer
	var errBuf *bytes.Buffer
	cmd.Stderr = imgg.stdErr
	if cmd.Stderr == nil {
		errBuf = new(bytes.Buffer)
		cmd.Stderr = errBuf
	}

	// set output to the desired writer or the internal buffer
	if imgg.outWriter != nil {
		cmd.Stdout = imgg.outWriter
	} else {
		imgg.outbuf.Reset() // reset internal buffer when we use it
		cmd.Stdout = &imgg.outbuf
	}

	//ioreader input
	if imgg.input != nil {
		cmd.Stdin = imgg.input
	}

	// run cmd to create the image
	err = cmd.Run()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		// on an error, return the error and the contents of Stderr if it was our own buffer
		// if Stderr was set to a custom writer, just return err
		if errBuf != nil {
			if errStr := errBuf.String(); strings.TrimSpace(errStr) != "" {
				return fmt.Errorf("%s\n%s", errStr, err)
			}
		}
		return err
	}
	return nil
}

// NewImageGenerator returns a new ImageGenerator struct with all options created and
// checks if wkhtmltoimage can be found on the system
func NewImageGenerator() (*ImageGenerator, error) {
	imgg := NewImagePreparer()
	return imgg, imgg.findPath()
}

// NewImagePreparer returns a ImageGenerator object without looking for the wkhtmltoimage executable file.
// This is useful to prepare a PDF file that is generated elsewhere and you just want to save the config as JSON.
// Note that Create() can not be called on this object unless you call SetImagePath yourself.
func NewImagePreparer() *ImageGenerator {
	return &ImageGenerator{
		imageOptions: newImageOptions(),
	}
}
