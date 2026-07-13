package cmd

import (
	"context"
	"io"
)

// GmailPDFExtractCmd is the hidden re-exec target behind `gmail attachment
// --text`: it reads PDF bytes on stdin and writes extracted text to stdout,
// so the parent can enforce a hard timeout by killing this process.
type GmailPDFExtractCmd struct{}

func (c *GmailPDFExtractCmd) Run(ctx context.Context, flags *RootFlags) error {
	text, err := pdfExtractChild(stdinReader(ctx))
	if err != nil {
		return err
	}
	_, err = io.WriteString(stdoutWriter(ctx), text)
	return err
}
