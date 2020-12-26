package ipvs

import "errors"

var errNotSupport = errors.New("not suuport")

// Client ipvs client
type Client struct {
}

// New returns a new ipvs client.
func New() (*Client, error) {
	return nil, errNotSupport
}

// AddService add a service.
func (c *Client) AddService(vs *Service) error {
	return errNotSupport
}

// DeleteService ...
func (c *Client) DeleteService(vs *Service) error {
	return errNotSupport
}

// GetServices ...
func (c *Client) GetServices() ([]*Service, error) {
	return nil, errNotSupport
}

// AddDestination ...
func (c *Client) AddDestination(vs *Service, rs *Destination) error {
	return errNotSupport
}

// DeleteDestination ...
func (c *Client) DeleteDestination(vs *Service, rs *Destination) error {
	return errNotSupport
}

// GetDestinations ...
func (c *Client) GetDestinations(vs *Service) ([]*Destination, error) {
	return nil, errNotSupport
}
