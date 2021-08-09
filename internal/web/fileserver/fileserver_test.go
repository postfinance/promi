package fileserver

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi"
)

func TestMuxFileServer(t *testing.T) {
	fixtures := map[string]http.File{
		"index.html": &testFile{"index.html", []byte("index\n")},
		"ok":         &testFile{"ok", []byte("ok\n")},
	}

	memfs := &testFileSystem{func(name string) (http.File, error) {
		name = name[1:]
		if f, ok := fixtures[name]; ok {
			return f, nil
		}
		return nil, errors.New("file not found")
	}}

	r := chi.NewRouter()
	FileServer(r, "/mounted", memfs)
	r.Get("/hi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bye"))
	})
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("nothing here"))
	})
	FileServer(r, "/", memfs)

	ts := httptest.NewServer(r)
	defer ts.Close()

	if _, body := testRequest(t, ts, "GET", "/hi", nil); body != "bye" {
		t.Fatalf(body)
	}

	// HEADS UP: net/http notfoundhandler will kick-in for static assets
	if _, body := testRequest(t, ts, "GET", "/mounted/nothing-here", nil); body == "nothing here" {
		t.Fatalf(body)
	}

	if _, body := testRequest(t, ts, "GET", "/nothing-here", nil); body == "nothing here" {
		t.Fatalf(body)
	}

	if _, body := testRequest(t, ts, "GET", "/mounted-nothing-here", nil); body == "nothing here" {
		t.Fatalf(body)
	}

	if _, body := testRequest(t, ts, "GET", "/hi", nil); body != "bye" {
		t.Fatalf(body)
	}

	if _, body := testRequest(t, ts, "GET", "/ok", nil); body != "ok\n" {
		t.Fatalf(body)
	}

	if _, body := testRequest(t, ts, "GET", "/mounted/ok", nil); body != "ok\n" {
		t.Fatalf(body)
	}

	// TODO/FIX: testFileSystem mock struct.. it struggles to pass this since it gets
	// into a redirect loop, however, it does work with http.Dir() using the disk.
	// if _, body := testRequest(t, ts, "GET", "/index.html", nil); body != "index\n" {
	// 	t.Fatalf(body)
	// }

	// if _, body := testRequest(t, ts, "GET", "/", nil); body != "index\n" {
	// 	t.Fatalf(body)
	// }

	// if _, body := testRequest(t, ts, "GET", "/mounted", nil); body != "index\n" {
	// 	t.Fatalf(body)
	// }

	// if _, body := testRequest(t, ts, "GET", "/mounted/", nil); body != "index\n" {
	// 	t.Fatalf(body)
	// }
}

type testFileSystem struct {
	open func(name string) (http.File, error)
}

func (fs *testFileSystem) Open(name string) (http.File, error) {
	return fs.open(name)
}

type testFile struct {
	name     string
	contents []byte
}

func (tf *testFile) Close() error {
	return nil
}

func (tf *testFile) Read(p []byte) (n int, err error) {
	copy(p, tf.contents)
	return len(p), nil
}

func (tf *testFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (tf *testFile) Readdir(count int) ([]os.FileInfo, error) {
	stat, _ := tf.Stat()
	return []os.FileInfo{stat}, nil
}

func (tf *testFile) Stat() (os.FileInfo, error) {
	return &testFileInfo{tf.name, int64(len(tf.contents))}, nil
}

type testFileInfo struct {
	name string
	size int64
}

func (tfi *testFileInfo) Name() string       { return tfi.name }
func (tfi *testFileInfo) Size() int64        { return tfi.size }
func (tfi *testFileInfo) Mode() os.FileMode  { return 0755 }
func (tfi *testFileInfo) ModTime() time.Time { return time.Now() }
func (tfi *testFileInfo) IsDir() bool        { return false }
func (tfi *testFileInfo) Sys() interface{}   { return nil }

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}
