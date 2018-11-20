package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/teris-io/cli"
	"github.com/xeipuuv/gojsonschema"
)

func doValidate(args []string, options map[string]string) int {
	//fmt.Println(options)
	verbose := options["verbose"] == "true"

	schemaAbsPath, err := filepath.Abs(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot get absolute path of schema\n")
		return 1
	}
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaAbsPath)

	jsonPath := args[1]
	fileName := filepath.Base(jsonPath)

	var reader io.Reader
	reader, err = os.Open(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open \"%s\": %s\n", jsonPath, err)
		return 1
	}

	if strings.HasSuffix(fileName, ".gz") {
		reader, err = gzip.NewReader(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot decompress %s: %s\n", jsonPath, err)
			return 1
		}
		fileName = fileName[:len(fileName)-3]
	}

	buf := bufio.NewReader(reader)

	if strings.HasSuffix(fileName, ".jsonl") {
		linenum := 1
		for {
			line, _, err := buf.ReadLine()
			if err == io.EOF {
				return 0
			}
			if err != nil {
				panic(err)
			}

			data := gojsonschema.NewStringLoader(string(line))

			result, err := gojsonschema.Validate(schemaLoader, data)

			if err != nil {
				panic(err.Error())
			}

			if result.Valid() {
				if verbose {
					fmt.Printf("line %d is valid\n", linenum)
				}
			} else {
				fmt.Printf("The document is not valid. see errors :\n")
				for _, desc := range result.Errors() {
					fmt.Printf("- %s\n", desc)
				}
			}

			linenum++
		}
	} else {
		data, _ := gojsonschema.NewReaderLoader(buf)

		result, err := gojsonschema.Validate(schemaLoader, data)
		if err != nil {
			panic(err.Error())
		}

		if result.Valid() {
			if verbose {
				fmt.Printf("The document is valid")
			}
		} else {
			fmt.Printf("The document is not valid. see errors :\n")
			for _, desc := range result.Errors() {
				fmt.Printf("- %s\n", desc)
			}
		}
	}

	return 0
}

func main() {
	app := cli.New("JSON Schema Validator @DEV@").
		WithOption(cli.NewOption("verbose", "Verbose mode").WithChar('v').WithType(cli.TypeBool)).
		WithArg(cli.NewArg("schema", "JSON Schema").WithType(cli.TypeString)).
		WithArg(cli.NewArg("json", "JSON file or JSON lines file (auto detect by suffix)").WithType(cli.TypeString)).
		WithAction(doValidate)

	os.Exit(app.Run(os.Args, os.Stdout))
}
