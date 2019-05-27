package router

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/dialogs/dialog-go-lib/service"
	"github.com/dialogs/dialog-go-lib/service/info"
	"github.com/stretchr/testify/require"
)

func TestAdminRouter(t *testing.T) {

	h, p := tempAddress(t)
	address := net.JoinHostPort(h, p)

	handleFunc := func(w http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	adminRouter := NewAdminRouter(&info.Info{
		Name:      "name",
		Version:   "version",
		Commit:    "commit",
		GoVersion: "goversion",
		BuildDate: "builddate",
	})
	adminRouter.HandleFunc("/custom", handleFunc)
	adminRouter.Handle("/sub", http.HandlerFunc(handleFunc))

	svc := service.NewHTTP(adminRouter, time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		require.Equal(t, http.ErrServerClosed, svc.ListenAndServeAddr(address))
	}()

	defer func() {
		require.NoError(t, svc.Close())
		wg.Wait()
	}()

	require.NoError(t, service.PingConn(address, 2, time.Second, nil))

	for _, testData := range []struct {
		Name string
		Fn   func(t *testing.T)
	}{
		{Name: "health", Fn: func(*testing.T) { testAdminRouterHandlerWithEmptyBody(t, address, "/health") }},
		{Name: "info", Fn: func(*testing.T) { testAdminRouterInfo(t, address) }},
		{Name: "custom", Fn: func(*testing.T) { testAdminRouterHandlerWithEmptyBody(t, address, "/custom") }},
		{Name: "sub", Fn: func(*testing.T) { testAdminRouterHandlerWithEmptyBody(t, address, "/sub") }},
	} {
		if !t.Run(testData.Name, testData.Fn) {
			return
		}
	}
}

func testAdminRouterHandlerWithEmptyBody(t *testing.T, address, path string) {

	endpoint := (&url.URL{
		Scheme: "http",
		Host:   address,
		Path:   path,
	}).String()

	// test: invalid method
	testInvalidMethod(t, endpoint)

	// test: ok
	res, err := http.Get(endpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer func() {
		require.NoError(t, res.Body.Close())
	}()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, []byte{}, body)
}

func testAdminRouterInfo(t *testing.T, address string) {

	endpoint := (&url.URL{
		Scheme: "http",
		Host:   address,
		Path:   "/info",
	}).String()

	// test: invalid method
	testInvalidMethod(t, endpoint)

	// test: ok
	res, err := http.Get(endpoint)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	defer func() {
		require.NoError(t, res.Body.Close())
	}()

	data := map[string]interface{}{}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&data))

	require.Equal(t,
		map[string]interface{}{
			"name":      "name",
			"version":   "version",
			"commit":    "commit",
			"goVersion": "goversion",
			"buildDate": "builddate",
		},
		data)
}

func testInvalidMethod(t *testing.T, endpoint string) {
	t.Helper()

	for _, method := range []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	} {

		req, err := http.NewRequest(method, endpoint, nil)
		require.NoError(t, err, endpoint)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err, endpoint)
		require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, endpoint)
	}
}

func tempAddress(t *testing.T) (host, port string) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	host, port, err = net.SplitHostPort(l.Addr().String())
	require.NoError(t, err)
	return
}
