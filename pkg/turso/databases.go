package turso

//
// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// )
//
// type DatabasesApi interface {
// 	Create(ctx context.Context, name, group string) (*CreateDatabaseResponse, error)
// }
//
// type Databases struct{
//
// }
//
// type DatabaseCreateResponse struct {
// 	Database struct {
// 		DbId     string `json:"DbId"`
// 		HostName string `json:"HostName"`
// 		Name     string `json:"Name"`
// 	} `json:"database"`
// }
//
// func (d *Databases) Create(ctx context.Context, name, group string) (*DatabaseResponse, error) {
// 	url := t.baseUrl + "/organizations/" + t.organization + "/databases"
// 	body := []byte(fmt.Sprintf(`{
// 		"name": "%s",
// 		"group": "%s"
// 	}`, name, group))
//
// 	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d))
//
// 	res, err := t.httpClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if res.StatusCode != 200 {
// 		var apiErr TursoError
//
// 		if err := json.NewDecoder(res.Body).Decode(&apiErr); err != nil {
// 			return nil, err
// 		}
//
// 		return nil, WrapErr(res.StatusCode, apiErr.Error)
// 	}
//
// 	var createdDatabase TursoDatabaseResponse
//
// 	if err := json.NewDecoder(res.Body).Decode(&createdDatabase); err != nil {
// 		return nil, err
// 	}
//
// 	return &createdDatabase, nil
//
// }
