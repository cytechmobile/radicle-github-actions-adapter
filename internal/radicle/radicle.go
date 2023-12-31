package radicle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"radicle-github-actions-adapter/app/radicle"
)

const patchURL string = "%s/api/v1/projects/%s/patches/%s"

type Radicle struct {
	nodeURL string
	token   string
	client  httpClient
	logger  *slog.Logger
}

func NewRadicle(nodeURL, token string, logger *slog.Logger) *Radicle {
	return &Radicle{
		nodeURL: nodeURL,
		token:   token,
		client:  http.DefaultClient,
		logger:  logger,
	}
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (r *Radicle) Comment(ctx context.Context, repoID, patchID, revisionID string, message string) error {
	payload := radicle.CreatePatchComment{
		Type:     radicle.CreatePatchCommentType,
		Body:     message,
		Revision: revisionID,
	}
	createRadiclePatchComment := fmt.Sprintf(patchURL, r.nodeURL, repoID, patchID)
	headers := map[string]string{}
	headers["content-type"] = "application/json"
	headers["Authorization"] = "Bearer " + r.token
	return r.request(ctx, createRadiclePatchComment, http.MethodPatch, headers, payload)

}

type HttpError struct {
	Status int
	Body   struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e HttpError) Error() string {
	return e.Body.Message
}

type HttpRequestError struct {
	Message string
}

func (e HttpRequestError) Error() string {
	return e.Message
}

func (r *Radicle) request(ctx context.Context, rawurl, method string, headers map[string]string,
	in interface{}) error {
	uri, err := url.Parse(rawurl)
	if err != nil {
		return err
	}

	// if we are posting or putting data, write them to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(in)
		if err != nil {
			return err
		}
	}

	// create a new http request.
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), buf)
	if err != nil {
		return err
	}
	for headerKey, headerVal := range headers {
		req.Header.Set(headerKey, headerVal)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// if an error is encountered, parse and return the error response.
	if resp.StatusCode >= http.StatusBadRequest {
		err := HttpError{}
		_ = json.NewDecoder(resp.Body).Decode(&err)
		err.Status = resp.StatusCode
		return err
	}

	return nil
}