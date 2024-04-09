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
	"strconv"
)

const patchURL string = "%s/api/v1/projects/%s/patches/%s"

type Radicle struct {
	nodeURL   string
	token     string
	client    httpClient
	logger    *slog.Logger
	commentID *string
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
		Embeds:   []string{},
	}
	if r.commentID != nil {
		payload.Type = radicle.EditPatchCommentType
		payload.Comment = r.commentID
	}
	createRadiclePatchComment := fmt.Sprintf(patchURL, r.nodeURL, repoID, patchID)
	headers := map[string]string{}
	headers["content-type"] = "application/json"
	headers["Authorization"] = "Bearer " + r.token
	type commentAddResp struct {
		Success bool   `json:"success"`
		Id      string `json:"id"`
	}
	resp := &commentAddResp{}
	err := r.request(ctx, createRadiclePatchComment, http.MethodPatch, headers, payload, resp)
	if nil == err && len(resp.Id) > 0 && r.commentID == nil {
		r.commentID = &resp.Id
	}
	return err
}

type HttpError struct {
	Status int
	Body   struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (e HttpError) Error() string {
	statusCode := ""
	if e.Status != 0 {
		statusCode = "HTTP" + strconv.Itoa(e.Status) + " "
	}
	return statusCode + e.Body.Message
}

func (r *Radicle) request(ctx context.Context, rawurl, method string, headers map[string]string,
	in, out interface{}) error {
	uri, err := url.Parse(rawurl)
	if err != nil {
		r.logger.Error("could not parse URL", "error", err.Error())
		return err
	}

	// if we are posting or putting data, write them to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(in)
		if err != nil {
			r.logger.Error("could not encode request payload", "error", err.Error())
			return err
		}
	}

	// create a new http request.
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), buf)
	if err != nil {
		r.logger.Error("could not invoke request", "error", err.Error())
		return err
	}
	for headerKey, headerVal := range headers {
		req.Header.Set(headerKey, headerVal)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		r.logger.Error("could not make request", "error", err.Error())
		return err
	}
	defer resp.Body.Close()

	// if an error is encountered, parse and return the error response.
	if resp.StatusCode >= http.StatusBadRequest {
		r.logger.Error("request responded with error code", "status code", resp.StatusCode)
		err := HttpError{}
		_ = json.NewDecoder(resp.Body).Decode(&err.Body.Message)
		err.Status = resp.StatusCode
		r.logger.Error("request responded with error message", "message", err.Body.Message)
		return err
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
