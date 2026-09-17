package client_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	http "github.com/bogdanfinn/fhttp"

	"github.com/solanab/nsk/internal/client"
)

var (
	errBoom   = errors.New("boom")
	errTLS    = errors.New("tls init")
	errAbs    = errors.New("abs")
	errStat   = errors.New("stat denied")
	errRead   = errors.New("read denied")
	errWrite  = errors.New("write denied")
	errChmod  = errors.New("chmod denied")
	errReq    = errors.New("bad request")
	errHTML   = errors.New("html parse")
	errEncode = errors.New("encode")
)

type errReader struct {
	err error
}

func (reader errReader) Read([]byte) (int, error) {
	return 0, reader.err
}

type closeBody struct {
	inner io.Reader
	err   error
}

func (body closeBody) Read(data []byte) (int, error) {
	return body.inner.Read(data) //nolint:wrapcheck // io.Reader must return raw EOF
}

func (body closeBody) Close() error {
	return body.err
}

func TestChromeHeaders(t *testing.T) {
	t.Parallel()

	var got http.Header

	doer := client.DoerFunc(func(req *http.Request) (*http.Response, error) {
		got = req.Header.Clone()
		if req.URL == nil || req.URL.String() != "https://"+forumHost+"/" {
			t.Fatalf("url = %v", req.URL)
		}

		if !requestHasPJWT(req) {
			t.Fatal("warmup request missing pjwt cookie")
		}

		return htmlResponse(http.StatusOK, fixture(t, htmlUser), nil), nil
	})

	_ = openOK(t, pjwtFile(t), doer)

	if headerVal(got, "user-agent") != wantUA {
		t.Fatalf("ua = %q", headerVal(got, "user-agent"))
	}

	if headerVal(got, "sec-ch-ua") != wantSecChUA {
		t.Fatalf("sec-ch-ua = %q", headerVal(got, "sec-ch-ua"))
	}
}

func TestDoError(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), client.DoerFunc(func(*http.Request) (*http.Response, error) {
		return nil, errBoom
	}))
	if !strings.Contains(err.Error(), "请求首页") {
		t.Fatalf("err = %v", err)
	}
}

func TestNilResponse(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), client.DoerFunc(func(*http.Request) (*http.Response, error) {
		var empty *http.Response

		return empty, nil
	}))
	if !strings.Contains(err.Error(), "空响应") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadBodyError(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), client.DoerFunc(func(*http.Request) (*http.Response, error) {
		resp := new(http.Response)
		resp.StatusCode = http.StatusOK
		resp.Header = http.Header{}
		resp.Body = closeBody{inner: errReader{err: errBoom}, err: nil}

		return resp, nil
	}))
	if !strings.Contains(err.Error(), "读取响应") {
		t.Fatalf("err = %v", err)
	}
}

func TestCloseBodyError(t *testing.T) {
	t.Parallel()

	err := openErr(t, pjwtFile(t), client.DoerFunc(func(*http.Request) (*http.Response, error) {
		resp := new(http.Response)
		resp.StatusCode = http.StatusOK
		resp.Header = http.Header{}
		resp.Body = closeBody{inner: strings.NewReader("<html></html>"), err: errBoom}

		return resp, nil
	}))
	if !strings.Contains(err.Error(), "关闭响应") {
		t.Fatalf("err = %v", err)
	}
}

func TestNewChromeClient(t *testing.T) {
	t.Parallel()

	if err := client.NewChromeClient(); err != nil {
		t.Fatal(err)
	}
}

func TestChromeTransportDo(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://www.nodeseek.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.ChromeTransport(userDoer(t)).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}

	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestChromeTransportError(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "https://www.nodeseek.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.ChromeTransport(client.DoerFunc(func(*http.Request) (*http.Response, error) {
		return nil, errBoom
	})).Do(req)
	if err == nil || !strings.Contains(err.Error(), "请求") {
		t.Fatalf("err = %v", err)
	}
}

func TestNilResponseHeaders(t *testing.T) {
	t.Parallel()

	_ = openOK(t, pjwtFile(t), client.DoerFunc(func(*http.Request) (*http.Response, error) {
		resp := new(http.Response)
		resp.StatusCode = http.StatusOK
		resp.Body = io.NopCloser(strings.NewReader(string(fixture(t, htmlUser))))

		return resp, nil
	}))
}

func TestHeaderNonEmptyFallback(t *testing.T) {
	t.Parallel()

	extra := http.Header{headerCF: {cfChallenge}}

	err := openErr(t, pjwtFile(t), fixtureDoer(t, htmlUser, http.StatusOK, extra))
	if !errors.Is(err, client.ErrCloudflare) {
		t.Fatalf("err = %v", err)
	}
}
