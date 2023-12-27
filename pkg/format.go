package pkg

import (
	"bytes"
	"mvdan.cc/sh/v3/syntax"
	"os"
)

func formatPkgbuild(file string) error {
	if data, err := os.ReadFile(file); err != nil {
		return err
	} else {
		if data, err := formatPkgbuildSingleLine(data); err != nil {
			return err
		} else if data, err := formatPkgbuildMultiLine(data); err != nil {
			return err
		} else if data, err := formatPkgbuildFinal(data); err != nil {
			return err
		} else if err := os.WriteFile(file, data, 0666); err != nil {
			return err
		}
	}

	return nil
}

func formatPkgbuildSingleLine(data []byte) ([]byte, error) {
	parser := syntax.NewParser(
		syntax.KeepComments(false),
		syntax.Variant(syntax.LangBash),
	)
	printer := syntax.NewPrinter(
		syntax.SingleLine(true),
	)

	return formatPkgbuildProcess(data, parser, printer)
}

func formatPkgbuildMultiLine(data []byte) ([]byte, error) {
	parser := syntax.NewParser(
		syntax.KeepComments(false),
		syntax.Variant(syntax.LangBash),
	)
	printer := syntax.NewPrinter(
		syntax.FunctionNextLine(true),
	)

	return formatPkgbuildProcess(data, parser, printer)
}

func formatPkgbuildFinal(data []byte) ([]byte, error) {
	parser := syntax.NewParser(
		syntax.KeepComments(false),
		syntax.Variant(syntax.LangBash),
	)
	printer := syntax.NewPrinter(
		syntax.Indent(0),
		syntax.BinaryNextLine(true),
		syntax.SwitchCaseIndent(true),
	)

	return formatPkgbuildProcess(data, parser, printer)
}

func formatPkgbuildProcess(data []byte, parser *syntax.Parser, printer *syntax.Printer) ([]byte, error) {
	if node, err := parser.Parse(bytes.NewReader(data), "PKGBUILD"); err != nil {
		return nil, err
	} else {
		buf := bytes.Buffer{}

		if err := printer.Print(&buf, node); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}
}
