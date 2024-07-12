package turso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TursoClient struct {
	organization string
	token        string
	httpClient   http.Client
	baseURL      string
}

type TursoError struct {
	Error string `json:"error"`
}

type TursoDatabase struct {
	DBID     string `json:"DbId"`
	HostName string `json:"HostName"`
	Name     string `json:"Name"`
}

type TursoDatabaseResponse struct {
	Database TursoDatabase `json:"database"`
}

type TursoDatabases struct {
	Databases []TursoDatabase `json:"databases"`
}

type TursoToken struct {
	Jwt string `json:"jwt"`
}

type TursoDatabaseAPI interface {
	CreateDatabase(name, group string) (*TursoDatabaseResponse, error)
	ListDatabases() (*TursoDatabases, error)
	CreateToken(name, expiration string) (*TursoToken, error)
	New(organization, apiToken string, httpClient http.Client) TursoClient
}

func (t TursoClient) New(organization, apiToken string, httpClient http.Client) TursoClient {
	return TursoClient{
		organization: organization,
		token:        apiToken,
		httpClient:   httpClient,
		baseURL:      "https://api.turso.tech/v1",
	}
}

func (t TursoClient) CreateDatabase(name, group string) (*TursoDatabaseResponse, error) {
	url := t.baseURL + "/organizations/" + t.organization + "/databases"
	body := []byte(fmt.Sprintf(`{
		"name": "%s",
		"group": "%s"
	}`, name, group))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))

	res, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var apiErr TursoError

		if err := json.NewDecoder(res.Body).Decode(&apiErr); err != nil {
			return nil, err
		}

		return nil, WrapErr(res.StatusCode, apiErr.Error)
	}

	var createdDatabase TursoDatabaseResponse

	if err := json.NewDecoder(res.Body).Decode(&createdDatabase); err != nil {
		return nil, err
	}

	return &createdDatabase, nil
}

func (t TursoClient) ListDatabases() (*TursoDatabases, error) {
	url := t.baseURL + "/organizations/" + t.organization + "/databases"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))

	res, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var databases TursoDatabases

	if err := json.NewDecoder(res.Body).Decode(&databases); err != nil {
		return nil, err
	}

	return &databases, nil
}

func (t TursoClient) CreateToken(name, expiration string) (*TursoToken, error) {
	url := t.baseURL + "/organizations/" + t.organization + "/databases/" + name + "/auth/tokens?expiration=" + expiration

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))

	res, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var token TursoToken

	if err := json.NewDecoder(res.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

type ErrConflict struct {
	StatusCode int
	Err        error
}

func (e ErrConflict) Error() string {
	return fmt.Sprint(e.Err.Error())
}

func WrapErr(statusCode int, msg string) error {
	switch statusCode {
	case http.StatusConflict:
		return ErrConflict{
			StatusCode: http.StatusConflict,
			Err:        fmt.Errorf(msg),
		}
	}

	return nil
}
