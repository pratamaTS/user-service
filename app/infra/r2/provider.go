package r2

func MustNew() *Client {
	c, err := New()
	if err != nil {
		panic(err)
	}
	return c
}
