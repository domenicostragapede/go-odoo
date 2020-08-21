// Package odoo exposes an XML-RPC client specifically designed for
// Odoo server (13, 12,11 officially supported).
//
// To start a connection with your Odoo server you have to build your
// ClientConfig data before using the NewClient method to start a new
// connection:
//
//		odooConfig := &odoo.ClientConfig{
//			Url:<odoo url path>,
//			Db:<database name>,
//			Username:<database user name>,
//			Password:<database password>
//		}
//
//		client, _ := odoo.NewClient(odooConfig)
//
// Note that NewClient will try to authenticate your config to the
// remote Url.
//
// Once authenticated, you can use the ExecuteKw method to perform a
// non ORM call (for an example a specific model call), otherwise you
// can use the standard ORM methods already wrapped.
package odoo

import "github.com/kolo/xmlrpc"

// Defines the args that must be passed to ExecuteKw method.
type Args []interface{}

// Attach one or more argument to the list
func (args *Args) Append(arg ...interface{}) {
	*args = append(*args, arg...)
}

// Define an abstract domain data structure.
// Domain can be created using the NewDomain function an the related Clause
// function for each clause included in domain.
// Example:
//		searchResult, err := client.Search("res.users", NewDomain(OpOR, Clause("active", "=", 1), Clause("login", "=", "John")))
type Domain []interface{}

// Defines the odoo domain operator AND
const OpAND = "&"

// Defines the odoo domain operator OR
const OpOR = "|"

// Defines a Odoo Clause as a tuple(field name, operator, value).
// Can be used as parameter in NewDomain function.
func Clause(field string, op string, values interface{}) interface{} {
	return []interface{}{field, op, values}
}

// Create a new Odoo Domain as a list of clause (OpAND and OpOR are admitted too).
func NewDomain(clauses ...interface{}) Domain {
	// TODO Consistence check
	return Domain{clauses}
}

// Client object will give you access to all remote methods.
type Client struct {
	cfg    *ClientConfig
	uid    int64
	auth   bool
	common *xmlrpc.Client
	object *xmlrpc.Client
}

// Check client authentication status.
func (client *Client) isAuthenticated() bool {
	return client.uid != 0 && client.auth
}

// Initialize Args list to be passed to ExecuteKw.
func (client *Client) getArgs() Args {
	var args Args
	if client.isAuthenticated() {
		args = Args{client.cfg.Db, client.uid, client.cfg.Password}
	} else {
		args = Args{client.cfg.Db, client.cfg.Username, client.cfg.Password, ""}
	}
	return args
}

// Abstract method to perform a XML-RPC call to an odoo server.
func (client *Client) call(c *xmlrpc.Client, method string, args []interface{}) (interface{}, error) {
	var result interface{}
	err := c.Call(method, args, &result)
	return result, err
}

// Perform XML-RPC Call to odoo using xmlrpc/2/common endpoint
func (client *Client) commonCall(method string, args Args) (interface{}, error) {
	return client.call(client.common, method, args)
}

// Perform XML-RPC Call to odoo using xmlrpc/2/object endpoint
func (client *Client) objectCall(method string, args Args) (interface{}, error) {
	return client.call(client.object, method, args)
}

// Perform remote authentication to the Odoo url defined into the ClientConfig.
func (client *Client) Authenticate() error {
	var err error = nil
	if !client.isAuthenticated() {
		if uid, err := client.commonCall("authenticate", client.getArgs()); uid != 0 && err == nil {
			client.uid = uid.(int64)
			client.auth = true
		} else {
			err = &ClientAuthError{config: client.cfg}
		}
	}
	return err
}

// Close the client connection.
func (client *Client) Close() error {
	// reset client data
	client.uid = 0
	client.auth = false

	var err error = nil
	if client.common != nil {
		err = client.common.Close()
	}

	if err == nil && client.object != nil {
		err = client.object.Close()
	}
	return err
}

// Call methods of odoo models via the execute_kw RPC function
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#calling-methods).
func (client *Client) ExecuteKw(method string, model string, args Args, context ...map[string]interface{}) (interface{}, error) {
	var err error = nil
	var result interface{} = nil
	var params = client.getArgs()
	params.Append(model, method, args)
	//for _, arg := range args {
	//	params.Append(arg)
	//}

	// context must be  0 or 1 value only
	if len(context) == 1 {
		params.Append(context[0])
	} else if len(context) >= 1 {
		err = &InvalidContextError{}
	}
	// if no context passed, this param will be ignored
	if err == nil {
		if client.isAuthenticated() {
			result, err = client.objectCall("execute_kw", params)
		} else {
			err = &ClientAuthError{config: client.cfg}
		}
	}
	return result, err
}

// Record data is accessible via the Read() method, which takes a list of ids
// (as returned by search()) and optionally a list of fields to fetch.
// By default, it will fetch all the fields the current user can read, which tends to be a huge amount
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#read-records).
func (client *Client) Read(model string, ids []int64, fields []string, context ...map[string]interface{}) (interface{}, error) {
	args := Args{ids}
	if len(fields) != 0 {
		args.Append(fields)
	}
	return client.ExecuteKw("read", model, args, context...)
}

// Records of a model are created using Create(). The method will create a single
// record and return its database identifier. Create() takes a mapping of fields
// to values, used to initialize the record. For any field which has a default
// value and is not set through the mapping argument, the default value will be used.
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#create-records).
func (client *Client) Create(model string, values map[string]interface{}, context ...map[string]interface{}) (int64, error) {
	var result int64 = 0
	response, err := client.ExecuteKw("create", model, Args{values}, context...)
	if err == nil {
		result = response.(int64)
	}
	return result, err
}

// Records can be updated using Write(). it takes a list of records to
// update and a mapping of updated fields to values similar to create()
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#update-records).
func (client *Client) Write(model string, ids []int64, values map[string]interface{}, context ...map[string]interface{}) (bool, error) {
	var result = false
	response, err := client.ExecuteKw("write", model, Args{ids, values}, context...)
	if err == nil {
		result = response.(bool)
	}
	return result, err
}

// Records can be deleted in bulk by providing their ids to Unlink()
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#delete-records).
func (client *Client) Unlink(model string, ids []int64, context ...map[string]interface{}) (bool, error) {
	var result = false
	response, err := client.ExecuteKw("unlink", model, Args{ids}, context...)
	if err == nil {
		result = response.(bool)
	}
	return result, err
}

// Records can be listed and filtered via Search().
// It takes a mandatory domain filter (possibly empty), and returns
// the database identifiers of all records matching the filter
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#list-records).
func (client *Client) Search(model string, domain Domain, context ...map[string]interface{}) ([]int64, error) {
	var result []int64 = nil
	var args = Args{}
	for _, clause := range domain {
		args.Append(clause)
	}
	response, err := client.ExecuteKw("search", model, args, context...)
	if err == nil {
		result = make([]int64, len(response.([]interface{})))
		for i, r := range response.([]interface{}) {
			result[i] = r.(int64)
		}
	}
	return result, err
}

// Odoo provides a SearchRead() shortcut which as its name suggests is equivalent
// to a Search() followed by a Read(), but avoids having to perform two requests
// and keep ids around. Its arguments are similar to Search()’s, but it can also
// take a list of fields (like Read(), if that list is not provided it will fetch
// all fields of matched records)
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#search-and-read).
func (client *Client) SearchRead(model string, domain Domain, fields []string, context ...map[string]interface{}) (interface{}, error) {
	args := Args{domain}
	if len(fields) != 0 {
		args.Append(fields)
	}
	return client.ExecuteKw("search_read", model, args, context...)
}

// SearchCount() can be used to retrieve only the number of records
// matching the query. It takes the same domain filter as search()
// and no other parameter
// (https://www.odoo.com/documentation/13.0/webservices/odoo.html#count-records).
func (client *Client) SearchCount(model string, domain Domain, context ...map[string]interface{}) (int64, error) {
	var result int64 = 0
	response, err := client.ExecuteKw("search_count", model, Args{domain}, context...)
	if err == nil {
		result = response.(int64)
	}
	return result, err
}

// Once authenticated, you can use the ExecuteKw method to perform a
// Initialize a new odoo client.
// Once the initialization is complete, a request of authentication
// will be performed.
func NewClient(config *ClientConfig) (*Client, error) {
	var err error = nil
	var client *Client = nil
	if config.IsValid() {
		var common, object *xmlrpc.Client
		common, err = xmlrpc.NewClient(config.Url+"/xmlrpc/2/common", nil)
		if err == nil {
			object, err = xmlrpc.NewClient(config.Url+"/xmlrpc/2/object", nil)
		}
		if err == nil {
			client = &Client{
				cfg:    config,
				uid:    0,
				auth:   false,
				common: common,
				object: object,
			}
			if err = client.Authenticate(); err != nil {
				client = nil
			}
		}
	} else {
		err = &InvalidConfigError{config: config}
	}
	return client, err
}
