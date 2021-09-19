package telego

import (
	stdJson "encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_apiResponse_String_and_APIError_Error(t *testing.T) {
	tests := []struct {
		name string
		resp apiResponse
		text string
	}{
		{
			name: "success",
			resp: apiResponse{
				Ok:       true,
				APIError: nil,
				Result:   stdJson.RawMessage{},
			},
			text: "Ok: true, Err: {<nil>}, Result: ",
		},
		{
			name: "error",
			resp: apiResponse{
				Ok: false,
				APIError: &APIError{
					Description: "bad request",
					ErrorCode:   400,
					Parameters:  nil,
				},
				Result: nil,
			},
			text: "Ok: false, Err: {400 \"bad request\"}, Result: ",
		},
		{
			name: "error with parameters",
			resp: apiResponse{
				Ok: false,
				APIError: &APIError{
					Description: "bad request",
					ErrorCode:   400,
					Parameters: &ResponseParameters{
						MigrateToChatID: 1,
						RetryAfter:      2,
					},
				},
				Result: nil,
			},
			text: "Ok: false, Err: {400 \"bad request\" migrate to chat id: 1, retry after: 2}, Result: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := tt.resp.String()
			assert.Equal(t, tt.text, text)
		})
	}
}
