# aws-get-web-identity-token

A single-binary CLI for the AWS STS [GetWebIdentityToken](https://docs.aws.amazon.com/STS/latest/APIReference/API_GetWebIdentityToken.html) API.

It returns a signed JSON Web Token (JWT) that represents the calling AWS identity, for use with external services that support OIDC discovery (IAM Outbound Identity Federation). This is the same operation as `aws sts get-web-identity-token`, packaged as a dependency-free static binary.

## Installation

```sh
brew install winebarrel/aws-get-web-identity-token/aws-get-web-identity-token
```

Or download a binary from the [releases page](https://github.com/winebarrel/aws-get-web-identity-token/releases), or build it yourself:

```sh
go install github.com/winebarrel/aws-get-web-identity-token/cmd/aws-get-web-identity-token@latest
```

## Usage

```
Usage: aws-get-web-identity-token --audience=AUDIENCE,... [flags]

Get a web identity token (JWT) from AWS STS GetWebIdentityToken.

Flags:
  -h, --help                       Show context-sensitive help.
      --version
  -a, --audience=AUDIENCE,...      Intended recipient of the token (aud claim).
                                   Repeat for multiple audiences.
  -d, --duration-seconds=INT-32    Token lifetime in seconds, 60-3600 (default
                                   300).
  -s, --signing-algorithm="RS256"  JWT signing algorithm: RS256 or ES384.
```

The token is printed to stdout:

```sh
$ aws-get-web-identity-token --audience my-service
eyJhbGciOiJSUzI1Ni...
```

## Authentication

Credentials and region are resolved with the default AWS SDK chain, so set them
via environment variables (or a shared profile):

```sh
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1
```

This command calls the STS API over the network; it does not generate the token
locally.
