package schemaregistry

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

const ContentType = "application/vnd.schemaregistry+json"

type Client struct {
	client  *http.Client
	uriPool sync.Pool
}

func NewClient(cfg IConfig) (*Client, error) {

	u, err := url.Parse(cfg.GetUrl())
	if err != nil {
		return nil, err
	}

	transport, err := cfg.GetTransport()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: cfg.GetTimeout(),
	}
	if transport != nil {
		client.Transport = transport
	}

	return &Client{
		client: client,
		uriPool: sync.Pool{
			New: func() interface{} {
				return &url.URL{
					Scheme: u.Scheme,
					Host:   u.Host,
					User:   u.User,
				}
			},
		},
	}, nil
}

// GetSchema :
// Get the schema string identified by the input id.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--schemas-ids-int-%20id
func (c *Client) GetSchema(ctx context.Context, id int) (*ResSchema, error) {

	u := c.getBaseURL()
	u.Path = "/schemas/ids/" + strconv.Itoa(id)

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResSchema{}
	if err := c.send(&res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// CheckSubject :
// Check if a schema has already been registered under the specified subject
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#post--subjects-(string-%20subject)
func (c *Client) CheckSubject(ctx context.Context, subject, schema string) (*ResCheckSubject, error) {

	u := c.getBaseURL()
	u.Path = "/subjects/" + url.PathEscape(subject)

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResCheckSubject{}
	if err := c.send(res, ctx, http.MethodPost, u, header, &ReqSubject{Schema: schema}); err != nil {
		return nil, err
	}

	return res, nil
}

// RegisterNewSchema :
// Register a new schema under the specified subject
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#post--subjects-(string-%20subject)-versions
func (c *Client) RegisterNewSchema(ctx context.Context, subject, schema string) (*ResRegisterNewSchema, error) {

	u := c.getBaseURL()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions"

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResRegisterNewSchema{}
	if err := c.send(res, ctx, http.MethodPost, u, header, &ReqSubject{Schema: schema}); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSubjectList :
// Get a list of registered subjects.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects
func (c *Client) GetSubjectList(ctx context.Context) (ResGetSubjectList, error) {

	u := c.getBaseURL()
	u.Path = "/subjects"

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := make(ResGetSubjectList, 0, 3)
	if err := c.send(&res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSubjectVersionsList :
// Get a list of versions registered under the specified subject.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects-(string-%20subject)-versions
func (c *Client) GetSubjectVersionsList(ctx context.Context, subject string) (ResGetSubjectVersionsList, error) {

	u := c.getBaseURL()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions"

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := make(ResGetSubjectVersionsList, 0, 3)
	if err := c.send(&res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// GetSubjectVersion :
// Get a specific version of the schema registered under this subject
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--subjects-(string-%20subject)-versions-(versionId-%20version)
func (c *Client) GetSubjectVersion(ctx context.Context, subject string, id int) (*ResGetSubjectVersion, error) {

	u := c.getBaseURL()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions/" + getID(id)

	header := http.Header{}
	header.Set("Accept", ContentType)

	res := &ResGetSubjectVersion{}
	if err := c.send(res, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteSubject :
// https://docs.confluent.io/current/schema-registry/schema-deletion-guidelines.html#schema-deletion-guidelines
func (c *Client) DeleteSubject(ctx context.Context, subject string) (ResDeleteSubject, error) {

	u := c.getBaseURL()
	u.Path = "/subjects/" + url.PathEscape(subject)

	res := make(ResDeleteSubject, 0, 3)
	if err := c.send(&res, ctx, http.MethodDelete, u, nil, nil); err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteSubjectVersion :
// https://docs.confluent.io/current/schema-registry/schema-deletion-guidelines.html#schema-deletion-guidelines
func (c *Client) DeleteSubjectVersion(ctx context.Context, subject string, id int) (ResDeleteSubjectVersion, error) {

	u := c.getBaseURL()
	u.Path = "/subjects/" + url.PathEscape(subject) + "/versions/" + getID(id)

	var res ResDeleteSubjectVersion
	if err := c.send(&res, ctx, http.MethodDelete, u, nil, nil); err != nil {
		return -1, err
	}

	return res, nil
}

// SetConfig :
// Update global compatibility level or update compatibility level for the specified subject.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#put--config
func (c *Client) SetConfig(ctx context.Context, subject string, cfg *ReqConfig) error {

	u := c.getBaseURL()
	u.Path = "/config"
	if len(subject) > 0 {
		u.Path += "/" + url.PathEscape(subject)
	}

	header := http.Header{}
	header.Set("Accept", ContentType)

	if err := c.send(nil, ctx, http.MethodPut, u, header, cfg); err != nil {
		return err
	}

	return nil
}

// GetConfig :
// Get global compatibility level.
// https://docs.confluent.io/2.0.1/schema-registry/docs/api.html#get--config
func (c *Client) GetConfig(ctx context.Context, subject string) (*ResConfig, error) {

	u := c.getBaseURL()
	u.Path = "/config"
	if len(subject) > 0 {
		u.Path += "/" + url.PathEscape(subject)
	}

	header := http.Header{}
	header.Set("Accept", ContentType)

	cfg := &ResConfig{}
	if err := c.send(cfg, ctx, http.MethodGet, u, header, nil); err != nil {
		return nil, err
	}

	return cfg, nil
}

// send request to a registry
func (c *Client) send(retval interface{}, ctx context.Context, method string, u *url.URL, header http.Header, payload interface{}) error {

	endpoint := u.String()
	c.putBaseURL(u)

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
			close(sendBodyErr)

		} else {
			reqBodyReader, reqBodyWriter := io.Pipe()
			defer reqBodyReader.Close()

			go func() {
				defer reqBodyWriter.Close()

				sendBodyErr <- json.NewEncoder(reqBodyWriter).Encode(payload)
			}()

			req.Body = reqBodyReader
		}
	}

	res, err := c.client.Do(req)
	if err == nil {
		if e, ok := <-sendBodyErr; ok {
			err = e
		}

		defer res.Body.Close()
	}

	if err != nil {
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

// getBaseURL returns base url from a pool
func (c *Client) getBaseURL() *url.URL {
	return c.uriPool.Get().(*url.URL)
}

// putBaseURL put url to a pool
func (c *Client) putBaseURL(u *url.URL) {

	u.Opaque = ""
	u.Path = ""
	u.RawPath = ""
	u.ForceQuery = false
	u.RawQuery = ""
	u.Fragment = ""

	c.uriPool.Put(u)
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
