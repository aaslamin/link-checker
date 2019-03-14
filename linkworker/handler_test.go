package linkworker_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aaslamin/link-checker/linkworker"
	"github.com/aaslamin/link-checker/storage"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {

	database := storage.NewMemoryStorage()
	workerHandler := linkworker.NewHandler(database)
	router := httprouter.New()
	workerHandler.SetupRoutes(router)

	type testState struct {
		workerRequest *http.Request
	}

	for _, testCase := range []struct {
		description string
		setup       func(t *testing.T, state *testState)
		assertions  func(t *testing.T, state *testState)
	}{
		{
			description: "should successfully create new worker and respond back with its id",
			setup: func(t *testing.T, state *testState) {
				reqBody := []byte(`{"url":"https://www.google.com"}`)
				req, err := http.NewRequest("POST", linkworker.WorkersHandlerPath, bytes.NewBuffer(reqBody))
				require.NoError(t, err)
				state.workerRequest = req
			},
			assertions: func(t *testing.T, state *testState) {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)

				assert.Equal(t, http.StatusCreated, recorder.Code)
				var workerResponse linkworker.Worker
				assert.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &workerResponse))
			},
		},
		{
			description: "should return 400 bad request if supplied URL is considered invalid",
			setup: func(t *testing.T, state *testState) {
				reqBody := []byte(`{"url":"this_is_not_a_valid_url"}`)
				req, err := http.NewRequest("POST", linkworker.WorkersHandlerPath, bytes.NewBuffer(reqBody))
				require.NoError(t, err)
				state.workerRequest = req
			},
			assertions: func(t *testing.T, state *testState) {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			description: "should return 400 bad request if request uses an invalid json scheme",
			setup: func(t *testing.T, state *testState) {
				reqBody := []byte(`{"totally_invalid" : "what_is_this"}`)
				req, err := http.NewRequest("POST", linkworker.WorkersHandlerPath, bytes.NewBuffer(reqBody))
				require.NoError(t, err)
				state.workerRequest = req
			},
			assertions: func(t *testing.T, state *testState) {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)
				assert.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			description: "should return 404 if worker cannot be found",
			setup: func(t *testing.T, state *testState) {
				workerId := "totally_not_a_valid_worker_id"

				req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", linkworker.WorkersHandlerPath, workerId), nil)
				require.NoError(t, err)
				state.workerRequest = req
			},
			assertions: func(t *testing.T, state *testState) {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			// in this test we first create a worker, parse the JSON response to fetch the ID
			// and then call the GET /workers endpoint to ensure we get a 200 back
			description: "should successfully be able to retrieve a worker using its ID",
			setup: func(t *testing.T, state *testState) {
				// step 1 - create worker by calling POST endpoint and retrieving ID from response
				reqBody := []byte(`{"url":"https://www.aporeto.com"}`)
				req, err := http.NewRequest("POST", linkworker.WorkersHandlerPath, bytes.NewBuffer(reqBody))
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				var workerResponse linkworker.Worker
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &workerResponse))
				require.NotEmpty(t, workerResponse.ID)

				// step 2 - create request to GET endpoint using the ID retrieved in step 1
				req, err = http.NewRequest("GET", fmt.Sprintf("%s/%s", linkworker.WorkersHandlerPath, workerResponse.ID), nil)
				require.NoError(t, err)
				state.workerRequest = req
			},
			assertions: func(t *testing.T, state *testState) {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)
				assert.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			description: "should return 404 if worker cannot be found on call to DELETE",
			setup: func(t *testing.T, state *testState) {
				workerId := "totally_not_a_valid_worker_id"

				req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", linkworker.WorkersHandlerPath, workerId), nil)
				require.NoError(t, err)
				state.workerRequest = req
			},
			assertions: func(t *testing.T, state *testState) {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			// in this test we first create a worker, parse the JSON response to fetch the ID
			// and then call the DELETE /workers endpoint to ensure we successfully deleted the worker
			description: "should successfully be able to retrieve a worker using its ID",
			setup: func(t *testing.T, state *testState) {
				// step 1 - create worker by calling POST endpoint and retrieving ID from response
				reqBody := []byte(`{"url":"https://www.aporeto.com"}`)
				req, err := http.NewRequest("POST", linkworker.WorkersHandlerPath, bytes.NewBuffer(reqBody))
				require.NoError(t, err)

				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, req)
				var workerResponse linkworker.Worker
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &workerResponse))
				require.NotEmpty(t, workerResponse.ID)

				// step 2 - create request to GET endpoint using the ID retrieved in step 1
				req, err = http.NewRequest("DELETE", fmt.Sprintf("%s/%s", linkworker.WorkersHandlerPath, workerResponse.ID), nil)
				require.NoError(t, err)
				state.workerRequest = req
			},
			assertions: func(t *testing.T, state *testState) {
				recorder := httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)
				assert.Equal(t, http.StatusOK, recorder.Code)

				// just to be sure, call the delete endpoint again, now it should be a 404 since we deleted the worker!
				recorder = httptest.NewRecorder()
				router.ServeHTTP(recorder, state.workerRequest)
				assert.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
	} {
		t.Run(fmt.Sprintf("scenario:%s", testCase.description), func(t *testing.T) {
			testSetupState := &testState{}
			testCase.setup(t, testSetupState)
			testCase.assertions(t, testSetupState)
		})
	}
}
