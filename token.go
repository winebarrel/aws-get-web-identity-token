package webidentity

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Context struct {
	Output io.Writer
	STS    *sts.Client
}

type Cmd struct {
	Audience         []string `short:"a" required:"" help:"Intended recipient of the token (aud claim). Repeat for multiple audiences."`
	DurationSeconds  int32    `short:"d" help:"Token lifetime in seconds, 60-3600 (default 300)."`
	SigningAlgorithm string   `short:"s" default:"RS256" enum:"RS256,ES384" help:"JWT signing algorithm: RS256 or ES384."`
}

func (cmd *Cmd) Run(cmdCtx *Context) error {
	input := &sts.GetWebIdentityTokenInput{
		Audience:         cmd.Audience,
		SigningAlgorithm: aws.String(cmd.SigningAlgorithm),
	}

	if cmd.DurationSeconds > 0 {
		input.DurationSeconds = aws.Int32(cmd.DurationSeconds)
	}

	out, err := cmdCtx.STS.GetWebIdentityToken(context.Background(), input)

	if err != nil {
		return err
	}

	fmt.Fprintln(cmdCtx.Output, aws.ToString(out.WebIdentityToken)) //nolint:errcheck

	return nil
}
