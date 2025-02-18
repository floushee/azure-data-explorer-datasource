package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/grafana/azure-data-explorer-datasource/pkg/azuredx/models"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	testUserLogin := "test-user"
	t.Run("When server returns 200", func(t *testing.T) {
		filename := "./testdata/successful-response.json"
		testDataRes, err := loadTestFile(filename)
		if err != nil {
			t.Errorf("test logic error: file doesnt exist: %s", filename)
		}
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusOK)
			_, err := rw.Write(testDataRes)
			if err != nil {
				t.Errorf("test logic error: %s", err.Error())
			}
		}))
		defer server.Close()

		payload := models.RequestPayload{
			DB:          "db-name",
			CSL:         "show databases",
			QuerySource: "schema",
		}

		client := New(server.Client())
		table, err := client.KustoRequest(server.URL, payload, nil)
		require.NoError(t, err)
		require.NotNil(t, table)
	})

	t.Run("When server returns 400", func(t *testing.T) {
		filename := "./testdata/error-response.json"
		testDataRes, err := loadTestFile(filename)
		if err != nil {
			t.Errorf("test logic error: file doesnt exist: %s", filename)
		}
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusBadRequest)
			_, err := rw.Write(testDataRes)
			if err != nil {
				t.Errorf("test logic error: %s", err.Error())
			}
		}))
		defer server.Close()

		payload := models.RequestPayload{
			DB:          "db-name",
			CSL:         "show databases",
			QuerySource: "schema",
		}

		client := New(server.Client())
		table, err := client.KustoRequest(server.URL, payload, nil)
		require.Nil(t, table)
		require.NotNil(t, err)
		require.Contains(t, err.Error(), "Request is invalid and cannot be processed: Syntax error: SYN0002: A recognition error occurred. [line:position=1:9]. Query: 'PerfTest take 5'")
	})

	t.Run("Headers are set", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			require.Equal(t, "application/json", req.Header.Get("Accept"))
			require.Equal(t, "application/json", req.Header.Get("Content-Type"))
			require.Equal(t, "Grafana-ADX", req.Header.Get("x-ms-app"))
			require.Contains(t, req.Header.Get("x-ms-client-request-id"), "KGC.schema")
			require.Contains(t, req.Header.Get("x-ms-user-id"), testUserLogin)
		}))
		defer server.Close()

		payload := models.RequestPayload{
			DB:          "db-name",
			CSL:         "show databases",
			QuerySource: "schema",
		}
		headers := map[string]string{
			"x-ms-user-id":           testUserLogin,
			"x-ms-client-request-id": fmt.Sprintf("KGC.%v;%v", "schema", uuid.Must(uuid.NewRandom()).String()),
		}

		client := New(server.Client())
		table, err := client.KustoRequest(server.URL, payload, headers)
		require.Nil(t, table)
		require.NotNil(t, err)
	})
}

func loadTestFile(path string) ([]byte, error) {
	jsonBody, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}
