package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	webidentity "github.com/winebarrel/aws-get-web-identity-token"
)

var version string

var cli struct {
	Version kong.VersionFlag
	webidentity.Cmd
}

func main() {
	kctx := kong.Parse(&cli,
		kong.Name("aws-get-web-identity-token"),
		kong.Description("Get a web identity token (JWT) from AWS STS GetWebIdentityToken."),
		kong.Vars{"version": version},
	)

	cfg, err := config.LoadDefaultConfig(context.Background())
	kctx.FatalIfErrorf(err)

	err = cli.Run(&webidentity.Context{
		Output: os.Stdout,
		STS:    sts.NewFromConfig(cfg),
	})
	kctx.FatalIfErrorf(err)
}
