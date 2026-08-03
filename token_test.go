package webidentity_test

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	webidentity "github.com/winebarrel/aws-get-web-identity-token"
)

const stsEndpoint = "https://sts.us-east-1.amazonaws.com/"

const testToken = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ0ZXN0In0.signature"

func tokenResponse(token string) string {
	return `<GetWebIdentityTokenResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">` +
		`<GetWebIdentityTokenResult>` +
		`<WebIdentityToken>` + token + `</WebIdentityToken>` +
		`<Expiration>2026-08-03T20:00:00Z</Expiration>` +
		`</GetWebIdentityTokenResult>` +
		`<ResponseMetadata><RequestId>req-id</RequestId></ResponseMetadata>` +
		`</GetWebIdentityTokenResponse>`
}

func errorResponse(code, message string) string {
	return `<ErrorResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">` +
		`<Error><Type>Sender</Type><Code>` + code + `</Code><Message>` + message + `</Message></Error>` +
		`<RequestId>req-id</RequestId></ErrorResponse>`
}

func newSTS(t *testing.T, opts ...func(*config.LoadOptions) error) (*sts.Client, *http.Client) {
	hc := &http.Client{}
	httpmock.ActivateNonDefault(hc)
	t.Cleanup(func() { httpmock.DeactivateNonDefault(hc) })

	opts = append(opts, config.WithHTTPClient(hc))
	cfg, err := config.LoadDefaultConfig(t.Context(), opts...)
	require.NoError(t, err)

	return sts.NewFromConfig(cfg), hc
}

func reqForm(t *testing.T, req *http.Request) url.Values {
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	form, err := url.ParseQuery(string(body))
	require.NoError(t, err)
	return form
}

func TestRun(t *testing.T) {
	stsClient, _ := newSTS(t)

	httpmock.RegisterResponder(http.MethodPost, stsEndpoint, func(req *http.Request) (*http.Response, error) {
		form := reqForm(t, req)
		require.Equal(t, "GetWebIdentityToken", form.Get("Action"))
		require.Equal(t, "my-service", form.Get("Audience.member.1"))
		require.Equal(t, "RS256", form.Get("SigningAlgorithm"))
		require.Equal(t, "900", form.Get("DurationSeconds"))
		return httpmock.NewStringResponse(http.StatusOK, tokenResponse(testToken)), nil
	})

	cmd := &webidentity.Cmd{
		Audience:         []string{"my-service"},
		DurationSeconds:  900,
		SigningAlgorithm: "RS256",
	}

	var buf bytes.Buffer
	err := cmd.Run(&webidentity.Context{STS: stsClient, Output: &buf})

	require.NoError(t, err)
	require.Equal(t, testToken+"\n", buf.String())
}

func TestRunMultipleAudiencesAndES384(t *testing.T) {
	stsClient, _ := newSTS(t)

	httpmock.RegisterResponder(http.MethodPost, stsEndpoint, func(req *http.Request) (*http.Response, error) {
		form := reqForm(t, req)
		require.Equal(t, "a", form.Get("Audience.member.1"))
		require.Equal(t, "b", form.Get("Audience.member.2"))
		require.Equal(t, "ES384", form.Get("SigningAlgorithm"))
		return httpmock.NewStringResponse(http.StatusOK, tokenResponse(testToken)), nil
	})

	cmd := &webidentity.Cmd{
		Audience:         []string{"a", "b"},
		SigningAlgorithm: "ES384",
	}

	err := cmd.Run(&webidentity.Context{STS: stsClient, Output: io.Discard})
	require.NoError(t, err)
}

func TestRunDefaultDuration(t *testing.T) {
	stsClient, _ := newSTS(t)

	httpmock.RegisterResponder(http.MethodPost, stsEndpoint, func(req *http.Request) (*http.Response, error) {
		form := reqForm(t, req)
		require.Empty(t, form.Get("DurationSeconds"))
		return httpmock.NewStringResponse(http.StatusOK, tokenResponse(testToken)), nil
	})

	cmd := &webidentity.Cmd{
		Audience:         []string{"my-service"},
		SigningAlgorithm: "RS256",
	}

	err := cmd.Run(&webidentity.Context{STS: stsClient, Output: io.Discard})
	require.NoError(t, err)
}

func TestRunError(t *testing.T) {
	stsClient, _ := newSTS(t, config.WithRetryer(func() aws.Retryer { return aws.NopRetryer{} }))

	httpmock.RegisterResponder(http.MethodPost, stsEndpoint, func(req *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusBadRequest, errorResponse("InvalidParameterValue", "bad audience")), nil
	})

	cmd := &webidentity.Cmd{
		Audience:         []string{"my-service"},
		SigningAlgorithm: "RS256",
	}

	err := cmd.Run(&webidentity.Context{STS: stsClient, Output: io.Discard})

	require.ErrorContains(t, err, "InvalidParameterValue")
	require.ErrorContains(t, err, "bad audience")
}
