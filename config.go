package odoo

// Define the base structure for connect to an odoo server.
type ClientConfig struct {
	Url      string
	Db       string
	Username string
	Password string
}

// Check if a configuration is valid.
func (config *ClientConfig) IsValid() bool {
	return config.Url != "" && config.Db != "" && config.Username != "" && config.Password != ""
}
