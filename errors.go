package odoo

import "fmt"

// This error will be returned when an authentication will fail.
type ClientAuthError struct {
	config *ClientConfig
	error
}

func (err *ClientAuthError) Error() string {
	return fmt.Sprintf(
		"Cannot authenticate to url %s on %s with user %s and password %s",
		err.config.Url, err.config.Db, err.config.Username, err.config.Password)
}

// This error will be returned when the odoo configuration is not valid
type InvalidConfigError struct {
	config *ClientConfig
	error
}

func (err *InvalidConfigError) Error() string {
	return fmt.Sprintf("Invalid Odoo configuration %s", err.config)
}

// This error will be returned when more than one context will be passed to a remote call.
type InvalidContextError struct {
	error
}

func (err *InvalidContextError) Error() string {
	return fmt.Sprintf("Maximum one context variable is admitted.")
}
