package client_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"

	"github.com/solanab/nsk/internal/client"
)

func TestNewTLSError(t *testing.T) {
	t.Cleanup(client.StubNewHTTPClient(func(
		_ tls_client.Logger,
		_ ...tls_client.HttpClientOption,
	) (tls_client.HttpClient, error) {
		return nil, errTLS
	}))

	_, err := client.New(pjwtFile(t))
	if err == nil || !strings.Contains(err.Error(), "创建 TLS 客户端失败") {
		t.Fatalf("err = %v", err)
	}
}

func TestAbsPathError(t *testing.T) {
	t.Cleanup(client.StubAbsPath(func(string) (string, error) { return "", errAbs }))

	err := openErr(t, "cookie.json", userDoer(t))
	if !strings.Contains(err.Error(), "cookie 路径") {
		t.Fatalf("err = %v", err)
	}
}

func TestStatError(t *testing.T) {
	t.Cleanup(client.StubStatFile(func(string) (os.FileInfo, error) { return nil, errStat }))

	err := openErr(t, pjwtFile(t), userDoer(t))
	if !strings.Contains(err.Error(), "读取 cookie 文件") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadFileError(t *testing.T) {
	t.Cleanup(client.StubReadFile(func(string) ([]byte, error) { return nil, errRead }))

	err := openErr(t, pjwtFile(t), userDoer(t))
	if !strings.Contains(err.Error(), "读取 cookie 文件") {
		t.Fatalf("err = %v", err)
	}
}

func TestWriteFileError(t *testing.T) {
	t.Cleanup(client.StubWriteFile(func(string, []byte, os.FileMode) error { return errWrite }))

	err := openErr(t, pjwtFile(t), userDoer(t))
	if !strings.Contains(err.Error(), "写入 cookie 文件") {
		t.Fatalf("err = %v", err)
	}
}

func TestChmodError(t *testing.T) {
	t.Cleanup(client.StubChmodFile(func(string, os.FileMode) error { return errChmod }))

	err := openErr(t, pjwtFile(t), userDoer(t))
	if !strings.Contains(err.Error(), "设置 cookie 文件权限") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewRequestError(t *testing.T) {
	t.Cleanup(client.StubNewRequest(func(
		_ context.Context,
		_, _ string,
		_ io.Reader,
	) (*http.Request, error) {
		return nil, errReq
	}))

	err := openErr(t, pjwtFile(t), userDoer(t))
	if !strings.Contains(err.Error(), "构造请求") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseHTMLError(t *testing.T) {
	t.Cleanup(client.StubParseHTML(func(io.Reader) (*goquery.Document, error) {
		return nil, errHTML
	}))

	err := openErr(t, pjwtFile(t), userDoer(t))
	if !strings.Contains(err.Error(), "pjwt 无效或过期") {
		t.Fatalf("err = %v", err)
	}
}

func TestMarshalIndentError(t *testing.T) {
	t.Cleanup(client.StubMarshalIndent(func(any, string, string) ([]byte, error) {
		return nil, errEncode
	}))

	err := openErr(t, pjwtFile(t), userDoer(t))
	if !strings.Contains(err.Error(), "编码 cookie JSON") {
		t.Fatalf("err = %v", err)
	}
}

func TestNilJarExport(t *testing.T) {
	t.Parallel()

	raw, err := client.NilJar().ExportCookiesJSON()
	if err != nil {
		t.Fatal(err)
	}

	if raw != "[]" {
		t.Fatalf("raw = %q", raw)
	}
}

func TestNonCanonicalCFHeader(t *testing.T) {
	t.Parallel()

	extra := http.Header{"CF-MITIGATED": {cfChallenge}}

	err := openErr(t, pjwtFile(t), fixtureDoer(t, htmlUser, http.StatusOK, extra))
	if err == nil || !strings.Contains(err.Error(), "cloudflare") {
		t.Fatalf("err = %v", err)
	}
}

func TestEmptyCFHeaderValue(t *testing.T) {
	t.Parallel()

	extra := http.Header{"CF-MITIGATED": {""}}
	forum := openOK(t, pjwtFile(t), fixtureDoer(t, htmlUser, http.StatusOK, extra))

	info, err := forum.WhoAmI()
	if err != nil {
		t.Fatal(err)
	}

	if info.Name != nameAlice {
		t.Fatalf("%#v", info)
	}
}
