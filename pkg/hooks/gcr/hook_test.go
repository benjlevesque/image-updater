package gcr

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bigkevmcd/image-hooks/pkg/hooks"
	"github.com/google/go-cmp/cmp"
)

var _ hooks.PushEvent = (*ContainerRegistryNotification)(nil)
var _ hooks.PushEventParser = Parse

func TestParseInsertWithoutTag(t *testing.T) {
	req := makeHookRequest(t, "testdata/gcr_insert.json")

	hook, err := Parse(req)
	if err != nil {
		t.Fatal(err)
	}
	if hook != nil {
		t.Fatalf("got %#v, want nil", hook)
	}
}

func TestParseInsertWithTag(t *testing.T) {
	req := makeHookRequest(t, "testdata/gcr_insert_with_tag.json")

	hook, err := Parse(req)
	if err != nil {
		t.Fatal(err)
	}

	want := &ContainerRegistryNotification{
		Action: "INSERT",
		Digest: "gcr.io/my-project/hello-world@sha256:6ec128e26cd5...",
		Tag:    "gcr.io/my-project/hello-world:1.1",
	}
	if diff := cmp.Diff(want, hook); diff != "" {
		t.Fatalf("hook doesn't match:\n%s", diff)
	}
}

func TestParseWithNoBody(t *testing.T) {
	bodyErr := errors.New("just a test error")

	req := httptest.NewRequest("POST", "/", failingReader{err: bodyErr})

	_, err := Parse(req)
	if err != bodyErr {
		t.Fatal("expected an error")
	}

}

func TestParseWithUnparseableBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)

	_, err := Parse(req)

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestPushedImageURL(t *testing.T) {
	hook := &ContainerRegistryNotification{
		Tag: "gcr.io/my-project/hello-world:1.1",
	}
	want := "gcr.io/my-project/hello-world:1.1"

	if u := hook.PushedImageURL(); u != want {
		t.Fatalf("got %s, want %s", u, want)
	}
}

func TestEventRepository(t *testing.T) {
	hook := &ContainerRegistryNotification{
		Digest: "gcr.io/my-project/hello-world@sha256:6ec128e26cd5...",
	}
	want := "my-project/hello-world"

	if u := hook.EventRepository(); u != want {
		t.Fatalf("got %s, want %s", u, want)
	}
}

func makeHookRequest(t *testing.T, fixture string) *http.Request {
	t.Helper()
	b, err := ioutil.ReadFile(fixture)
	if err != nil {
		t.Fatalf("failed to read %s: %s", fixture, err)
	}
	req := httptest.NewRequest("POST", "/", bytes.NewReader(b))
	req.Header.Add("Content-Type", "application/json")
	return req
}

type failingReader struct {
	err error
}

func (f failingReader) Read(p []byte) (n int, err error) {
	return 0, f.err
}
func (f failingReader) Close() error {
	return f.err
}
