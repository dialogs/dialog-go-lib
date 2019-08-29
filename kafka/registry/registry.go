package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const ContentType = "application/vnd.schemaregistry+json"

type Registry struct {
	client  *http.Client
	uriPool sync.Pool
}

func New(uri string, timeout time.Duration, certPem []byte) (*Registry, error) {

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: timeout,
	}

	if len(certPem) > 0 {
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(certPem) {
			return nil, errors.New("invalid pem data")
		}

		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caPool,
			},
		}
	}

	return &Registry{
		client: client,
		uriPool: sync.Pool{
			New: func() interface{} {
				return &url.URL{
					Scheme: u.Scheme,
					Host:   u.Host,
				}
			},
		},
	}, nil
}

// GetSchema :
// Get the schema string identified by the input id.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--schemas-ids-int-%20id
func (r *Registry) GetSchema(ctx context.Context, id int) (*ResSchema, error) {

	u := r.getBaseUrl()
	u.Path = "/schemas/ids/" + strconv.Itoa(id)

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResSchema{}
	if err := r.send(&res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// CheckSubject :
// Check if a schema has already been registered under the specified subject
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#post--subjects-(string-%20subject)
func (r *Registry) CheckSubject(ctx context.Context, subject, schema string) (*ResCheckSubject, error) {

	u := r.getBaseUrl()
	u.Path = "/subjects/" + url.PathEscape(subject)

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResCheckSubject{}
	if err := r.send(res, ctx, http.MethodPost, u, header, &ReqSubject{Schema: schema}); err != nil {
		return nil, err
	}

	return res, nil
}

// RegisterNewSchema :
// Register a new schema under the specified subject
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#post--subjects-(string-%20subject)-versions
func (r *Registry) RegisterNewSchema(ctx context.Context, subject, schema string) (*ResRegisterNewSchema, error) {

	u := r.getBaseUrl()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions"

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResRegisterNewSchema{}
	if err := r.send(res, ctx, http.MethodPost, u, header, &ReqSubject{Schema: schema}); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSubjectList :
// Get a list of registered subjects.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects
func (r *Registry) GetSubjectList(ctx context.Context) (ResGetSubjectList, error) {

	u := r.getBaseUrl()
	u.Path = "/subjects"

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := make(ResGetSubjectList, 0, 3)
	if err := r.send(&res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSubjectVersionsList :
// Get a list of versions registered under the specified subject.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects-(string-%20subject)-versions
func (r *Registry) GetSubjectVersionsList(ctx context.Context, subject string) (ResGetSubjectVersionsList, error) {

	u := r.getBaseUrl()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions"

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := make(ResGetSubjectVersionsList, 0, 3)
	if err := r.send(&res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSubjectVersion :
// Get a specific version of the schema registered under this subject
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects-(string-%20subject)-versions-(versionId-%20version)
func (r *Registry) GetSubjectVersion(ctx context.Context, subject string, id int) (*ResGetSubjectVersion, error) {

	u := r.getBaseUrl()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions/" + getID(id)

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResGetSubjectVersion{}
	if err := r.send(res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteSubject :
// https://docs.confluent.io/current/schema-registry/schema-deletion-guidelines.html#schema-deletion-guidelines
func (r *Registry) DeleteSubject(ctx context.Context, subject string) (ResDeleteSubject, error) {

	u := r.getBaseUrl()
	u.Path = "/subjects/" + url.PathEscape(subject)

	res := make(ResDeleteSubject, 0, 3)
	if err := r.send(&res, ctx, http.MethodDelete, u, nil, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteSubjectVersion :
// https://docs.confluent.io/current/schema-registry/schema-deletion-guidelines.html#schema-deletion-guidelines
func (r *Registry) DeleteSubjectVersion(ctx context.Context, subject string, id int) (ResDeleteSubjectVersion, error) {

	u := r.getBaseUrl()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions/" + getID(id)

	var res ResDeleteSubjectVersion
	if err := r.send(&res, ctx, http.MethodDelete, u, nil, nil); err != nil {
		return -1, err
	}

	return res, nil
}

// SetConfig :
// Update global compatibility level or update compatibility level for the specified subject.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#put--config
func (r *Registry) SetConfig(ctx context.Context, subject string, cfg *ReqConfig) error {

	u := r.getBaseUrl()
	u.Path = "/config"
	if len(subject) > 0 {
		u.Path += "/" + url.PathEscape(subject)
	}

	header := http.Header{}
	header.Set("Accept", ContentType)

	if err := r.send(nil, ctx, http.MethodPut, u, header, cfg); err != nil {
		return err
	}

	return nil
}

// GetConfig :
// Get global compatibility level.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--config
func (r *Registry) GetConfig(ctx context.Context, subject string) (*ResConfig, error) {

	u := r.getBaseUrl()
	u.Path = "/config"
	if len(subject) > 0 {
		u.Path += "/" + url.PathEscape(subject)
	}

	header := http.Header{}
	header.Set("Accept", ContentType)

	cfg := &ResConfig{}
	if err := r.send(cfg, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return cfg, nil
}

// send request to a registry
func (r *Registry) send(retval interface{}, ctx context.Context, method string, u *url.URL, header http.Header, payload interface{}) error {

	endpoint := u.String()
	r.putBaseUrl(u)

	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return err
	}

	sendBodyErr := make(chan error, 1)
	{
		// prepare http request
		req.Header = header
		req = req.WithContext(ctx)

		if payload == nil {
			sendBodyErr <- nil
			defer close(sendBodyErr)

		} else {
			reqBodyReader, reqBodyWriter := io.Pipe()
			defer reqBodyReader.Close()

			go func() {
				defer reqBodyWriter.Close()
				defer close(sendBodyErr)

				sendBodyErr <- json.NewEncoder(reqBodyWriter).Encode(payload)
			}()

			req.Body = reqBodyReader
			req.GetBody = func() (io.ReadCloser, error) {
				return reqBodyReader, nil
			}
		}
	}

	res, err := r.client.Do(req)
	if err == nil {
		err = <-sendBodyErr
		defer res.Body.Close()
	}

	if err != nil {
		res.Body.Close()
		return err
	}

	if res.StatusCode != http.StatusOK {

		errInfo := newError(res.StatusCode)
		if err := json.NewDecoder(res.Body).Decode(errInfo); err != nil {
			return errors.Wrap(err, "failed to decode body for status:"+strconv.Itoa(res.StatusCode))
		}

		return errInfo
	}

	if retval != nil {
		if err := json.NewDecoder(res.Body).Decode(retval); err != nil {
			return err
		}
	}

	return nil
}

// getBaseUrl returns base url from a pool
func (r *Registry) getBaseUrl() *url.URL {
	return r.uriPool.Get().(*url.URL)
}

// putBaseUrl put url to a pool
func (r *Registry) putBaseUrl(u *url.URL) {

	u.Opaque = ""
	u.User = nil
	u.Path = ""
	u.RawPath = ""
	u.ForceQuery = false
	u.RawQuery = ""
	u.Fragment = ""

	r.uriPool.Put(u)
}

// getID returns id for url
func getID(src int) (id string) {

	if src == -1 {
		id = "latest"
	} else {
		id = strconv.Itoa(src)
	}

	return
}
