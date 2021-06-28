// Code generated by go-swagger; DO NOT EDIT.

package bluemix_service_instances

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new bluemix service instances API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for bluemix service instances API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
BluemixServiceInstanceGet gets the current state information associated with the service instance
*/
func (a *Client) BluemixServiceInstanceGet(params *BluemixServiceInstanceGetParams, authInfo runtime.ClientAuthInfoWriter) (*BluemixServiceInstanceGetOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewBluemixServiceInstanceGetParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "bluemix.serviceInstance.get",
		Method:             "GET",
		PathPattern:        "/bluemix_v1/service_instances/{instance_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &BluemixServiceInstanceGetReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*BluemixServiceInstanceGetOK), nil

}

/*
BluemixServiceInstancePut updates disable or enable the state of a provisioned service instance
*/
func (a *Client) BluemixServiceInstancePut(params *BluemixServiceInstancePutParams, authInfo runtime.ClientAuthInfoWriter) (*BluemixServiceInstancePutOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewBluemixServiceInstancePutParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "bluemix.serviceInstance.put",
		Method:             "PUT",
		PathPattern:        "/bluemix_v1/service_instances/{instance_id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &BluemixServiceInstancePutReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return result.(*BluemixServiceInstancePutOK), nil

}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
