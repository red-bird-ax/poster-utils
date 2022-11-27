package network

import (
    "encoding/json"
    "net/http"
)

func UnmarshalRequestBody[T any](request *http.Request) (*T, error) {
    var requestBody T
    if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
        return nil, err
    }
    _ = request.Body.Close()
    return &requestBody, nil
}

func ResponseError(response http.ResponseWriter, message string, code int) {
    responseBody := struct {
        Message string `json:"message"`
        Error   string `json:"error"`
    }{
        Message: message,
        Error:   http.StatusText(code),
    }
    response.WriteHeader(code)
    _ = json.NewEncoder(response).Encode(&responseBody)
}