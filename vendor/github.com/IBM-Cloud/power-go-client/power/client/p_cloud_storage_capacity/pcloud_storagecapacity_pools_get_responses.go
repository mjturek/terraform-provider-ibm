// Code generated by go-swagger; DO NOT EDIT.

package p_cloud_storage_capacity

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/IBM-Cloud/power-go-client/power/models"
)

// PcloudStoragecapacityPoolsGetReader is a Reader for the PcloudStoragecapacityPoolsGet structure.
type PcloudStoragecapacityPoolsGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PcloudStoragecapacityPoolsGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 200:
		result := NewPcloudStoragecapacityPoolsGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 401:
		result := NewPcloudStoragecapacityPoolsGetUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 404:
		result := NewPcloudStoragecapacityPoolsGetNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	case 500:
		result := NewPcloudStoragecapacityPoolsGetInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewPcloudStoragecapacityPoolsGetOK creates a PcloudStoragecapacityPoolsGetOK with default headers values
func NewPcloudStoragecapacityPoolsGetOK() *PcloudStoragecapacityPoolsGetOK {
	return &PcloudStoragecapacityPoolsGetOK{}
}

/*PcloudStoragecapacityPoolsGetOK handles this case with default header values.

OK
*/
type PcloudStoragecapacityPoolsGetOK struct {
	Payload *models.StoragePoolCapacity
}

func (o *PcloudStoragecapacityPoolsGetOK) Error() string {
	return fmt.Sprintf("[GET /pcloud/v1/cloud-instances/{cloud_instance_id}/storage-capacity/storage-pools/{storage_pool_name}][%d] pcloudStoragecapacityPoolsGetOK  %+v", 200, o.Payload)
}

func (o *PcloudStoragecapacityPoolsGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.StoragePoolCapacity)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPcloudStoragecapacityPoolsGetUnauthorized creates a PcloudStoragecapacityPoolsGetUnauthorized with default headers values
func NewPcloudStoragecapacityPoolsGetUnauthorized() *PcloudStoragecapacityPoolsGetUnauthorized {
	return &PcloudStoragecapacityPoolsGetUnauthorized{}
}

/*PcloudStoragecapacityPoolsGetUnauthorized handles this case with default header values.

Unauthorized
*/
type PcloudStoragecapacityPoolsGetUnauthorized struct {
	Payload *models.Error
}

func (o *PcloudStoragecapacityPoolsGetUnauthorized) Error() string {
	return fmt.Sprintf("[GET /pcloud/v1/cloud-instances/{cloud_instance_id}/storage-capacity/storage-pools/{storage_pool_name}][%d] pcloudStoragecapacityPoolsGetUnauthorized  %+v", 401, o.Payload)
}

func (o *PcloudStoragecapacityPoolsGetUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPcloudStoragecapacityPoolsGetNotFound creates a PcloudStoragecapacityPoolsGetNotFound with default headers values
func NewPcloudStoragecapacityPoolsGetNotFound() *PcloudStoragecapacityPoolsGetNotFound {
	return &PcloudStoragecapacityPoolsGetNotFound{}
}

/*PcloudStoragecapacityPoolsGetNotFound handles this case with default header values.

Not Found
*/
type PcloudStoragecapacityPoolsGetNotFound struct {
	Payload *models.Error
}

func (o *PcloudStoragecapacityPoolsGetNotFound) Error() string {
	return fmt.Sprintf("[GET /pcloud/v1/cloud-instances/{cloud_instance_id}/storage-capacity/storage-pools/{storage_pool_name}][%d] pcloudStoragecapacityPoolsGetNotFound  %+v", 404, o.Payload)
}

func (o *PcloudStoragecapacityPoolsGetNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewPcloudStoragecapacityPoolsGetInternalServerError creates a PcloudStoragecapacityPoolsGetInternalServerError with default headers values
func NewPcloudStoragecapacityPoolsGetInternalServerError() *PcloudStoragecapacityPoolsGetInternalServerError {
	return &PcloudStoragecapacityPoolsGetInternalServerError{}
}

/*PcloudStoragecapacityPoolsGetInternalServerError handles this case with default header values.

Internal Server Error
*/
type PcloudStoragecapacityPoolsGetInternalServerError struct {
	Payload *models.Error
}

func (o *PcloudStoragecapacityPoolsGetInternalServerError) Error() string {
	return fmt.Sprintf("[GET /pcloud/v1/cloud-instances/{cloud_instance_id}/storage-capacity/storage-pools/{storage_pool_name}][%d] pcloudStoragecapacityPoolsGetInternalServerError  %+v", 500, o.Payload)
}

func (o *PcloudStoragecapacityPoolsGetInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
