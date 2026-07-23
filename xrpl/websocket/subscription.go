package websocket

import subscribe "github.com/Peersyst/xrpl-go/xrpl/queries/subscription"

// Subscribe subscribes to the streams and accounts specified in the request.
// It returns a response from the server.
func (c *Client) Subscribe(req *subscribe.Request) (*subscribe.Response, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr subscribe.Response
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}

// Unsubscribe unsubscribes from the streams and accounts specified in the request.
// It returns a response from the server.
func (c *Client) Unsubscribe(req *subscribe.UnsubscribeRequest) (*subscribe.UnsubscribeResponse, error) {
	res, err := c.Request(req)
	if err != nil {
		return nil, err
	}
	var lr subscribe.UnsubscribeResponse
	err = res.GetResult(&lr)
	if err != nil {
		return nil, err
	}
	return &lr, nil
}
