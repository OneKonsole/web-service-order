package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	okmq "github.com/OneKonsole/sys-queueing"

	amqp "github.com/rabbitmq/amqp091-go"
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
func produce(
	channel *amqp.Channel,
	routingKey string,
	exchangeName string,
	contentType string,
	maxProcessTime int,
	messageBody []byte) {

	// Messages from this producer must not take more than 5 seconds to be produced
	channelContext, cancel := context.WithTimeout(context.Background(), time.Duration(maxProcessTime)*time.Second)
	defer cancel()

	err := channel.PublishWithContext(channelContext,
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  contentType,
			Body:         messageBody,
		})

	okmq.FailOnError(err, "Failed to produce messages for provisioning")
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, req)

	return recorder
}
