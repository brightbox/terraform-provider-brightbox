package brightbox

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	brightbox "github.com/brightbox/gobrightbox/v2"
)

// brightboxAPIRequest sends a raw HTTP request for update paths whose JSON
// body gobrightbox's typed option structs can't represent - see
// requestBodyWithNullField. Not a general-purpose REST client: use the
// generated gobrightbox client methods for everything else.
func brightboxAPIRequest[O any](ctx context.Context, client *brightbox.Client, method string, relURL string, body interface{}) (*O, error) {
	absURL, err := client.ResourceBaseURL().Parse(relURL)
	if err != nil {
		return nil, err
	}
	// Parse drops the base URL's query (e.g. the account param); carry it
	// over explicitly.
	absURL.RawQuery = client.ResourceBaseURL().RawQuery

	var buf bytes.Buffer
	if body != nil {
		err = json.NewEncoder(&buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, absURL.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if client.UserAgent != "" {
		req.Header.Add("User-Agent", client.UserAgent)
	}

	res, err := client.HTTPClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		apiErr := &brightbox.APIError{
			RequestURL: res.Request.URL,
			StatusCode: res.StatusCode,
			Status:     res.Status,
		}
		apiErr.ResponseBody, apiErr.ParseError = io.ReadAll(res.Body)
		if len(apiErr.ResponseBody) > 0 {
			apiErr.ParseError = json.Unmarshal(apiErr.ResponseBody, apiErr)
		}
		return nil, apiErr
	}

	result := new(O)
	err = json.NewDecoder(res.Body).Decode(result)
	if err != nil {
		return nil, &brightbox.APIError{
			RequestURL: res.Request.URL,
			StatusCode: res.StatusCode,
			Status:     res.Status,
			ParseError: err,
		}
	}
	return result, nil
}

// requestBodyWithNullField marshals body to JSON, then overwrites the given
// fields with an explicit JSON null. gobrightbox option structs use
// `*string` with `omitempty`, which only checks pointer nil-ness: a
// non-nil pointer to "" still marshals to `""`, never `null`. This is the
// only way to send a real null without changing the gobrightbox types.
func requestBodyWithNullField(body interface{}, nullFields ...string) (map[string]interface{}, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	var requestBody map[string]interface{}
	err = json.Unmarshal(encoded, &requestBody)
	if err != nil {
		return nil, err
	}
	for _, field := range nullFields {
		requestBody[field] = nil
	}
	return requestBody, nil
}
