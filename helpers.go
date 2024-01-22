package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

// ===========================================================================================================
// Helper to create a HTTP error message. The message will be sent as JSON
// Parameters:
//
//	w (http.ResponseWriter) : Helper object to create HTTP responses
//	code (int) : HTTP code to send
//	message (string) : Error message to send
//
// Examples:
//
//	respondWithError(w, 500, "Couldn't process the order")
//
// ===========================================================================================================
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// ===========================================================================================================
// Helper to create JSON HTTP responses
// Parameters:
//
//	w (http.ResponseWriter) : Helper object to create HTTP responses
//	code (int) : HTTP code to send
//	payload (interface) : Data to answer with
//
// Examples:
//
//	respondWithJSON(w, 200, new Order(xx,xx,xx,xx)")
//
// ===========================================================================================================
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// ===========================================================================================================
// Facilitate messages producing with RabbitMQ.
//
// Parameters:
//
//	processTime (int): Expected time taken to produce the message
//	channel (*amqp.Channel): Channel in which one to produce
//	dialName (string): URI of the RabbitMQ connection
//	exchangeName (string) : Name of the exchange where to send the message
//	contentType (string) : HTTP like ContentType (e.g. text/plain)
//	messageBody ([]byte) : Message to send to the queue
//
// Examples:
//
//	produce(&channel, "xxx", "xxx", "application/json", 5 , byte_array_containing_var)
//
// ===========================================================================================================

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, req)

	return recorder
}

func isValidClusterName(fl validator.FieldLevel) bool {
	// Define the regular expression pattern
	clusterNamePattern := "^[a-z0-9][a-z0-9-]*[a-z0-9]$"

	// Compile the regular expression
	regex := regexp.MustCompile(clusterNamePattern)

	// Extract the field value
	clusterName := fl.Field().String()

	// Check if the clusterName matches the pattern
	return regex.MatchString(clusterName)
}

func startsWithAlphanum(fl validator.FieldLevel) bool {
	firstChar, _ := utf8.DecodeLastRuneInString(fl.Field().String())
	return unicode.IsLetter(firstChar) || unicode.IsDigit(firstChar)
}

func endWithAlphanum(fl validator.FieldLevel) bool {
	lastChar, _ := utf8.DecodeLastRuneInString(fl.Field().String())
	return unicode.IsLetter(lastChar) || unicode.IsDigit(lastChar)
}
func isUUID(fl validator.FieldLevel) bool {
	// Define the expected UUID format using a regular expression
	uuidPattern := regexp.MustCompile(`^[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}$`)

	// Check if the field matches the expected UUID format
	return uuidPattern.MatchString(fl.Field().String())
}
