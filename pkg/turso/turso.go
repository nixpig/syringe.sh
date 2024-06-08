package turso

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type TursoApi struct {
	organization string
	apiToken     string
	httpClient   http.Client
	baseUrl      string
}

type TursoApiError struct {
	Error string `json:"error"`
}

type TursoDatabase struct {
	DbId     string `json:"DbId"`
	HostName string `json:"HostName"`
	Name     string `json:"Name"`
}

type TursoDatabases struct {
	Databases []TursoDatabase `json:"databases"`
}

type TursoDatabaseApi interface {
	CreateDatabase(name, group string) (*TursoDatabase, error)
	ListDatabases() (*[]TursoDatabase, error)
}

func New(organization, apiToken string, httpClient http.Client) TursoApi {
	return TursoApi{
		organization: organization,
		apiToken:     apiToken,
		httpClient:   httpClient,
		baseUrl:      "https://api.turso.tech/v1",
	}
}

func (t *TursoApi) CreateDatabase(name, group string) (*TursoDatabase, error) {
	url := t.baseUrl + "/organizations/" + t.organization + "/databases"
	body := []byte(fmt.Sprintf(`{
		"name": "%s",
		"group": "%s"
	}`, name, group))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiToken))

	res, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		var apiErr TursoApiError

		if err := json.NewDecoder(res.Body).Decode(&apiErr); err != nil {
			return nil, err
		}

		return nil, WrapErr(res.StatusCode, apiErr.Error)
	}

	var createdDatabase TursoDatabase

	if err := json.NewDecoder(res.Body).Decode(&createdDatabase); err != nil {
		return nil, err
	}

	return &createdDatabase, nil
}

func (t *TursoApi) ListDatabases() (*TursoDatabases, error) {
	url := t.baseUrl + "/organizations/" + t.organization + "/databases"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiToken))

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

type ErrConflict struct {
	StatusCode int
	Err        error
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf(e.Err.Error())
}

func WrapErr(statusCode int, msg string) error {
	switch statusCode {
	case http.StatusConflict:
		return ErrConflict{
			StatusCode: http.StatusConflict,
			Err:        errors.New(msg),
		}
	}

	return nil
}
