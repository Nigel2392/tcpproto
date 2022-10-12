package tcpproto

type Cookies struct {
	// Cookies are stored in a map
	// The key is the name of the cookie
	// The value is the cookie itself
	cookies map[string]*Cookie
}

func InitCookies() *Cookies {
	return &Cookies{
		cookies: make(map[string]*Cookie),
	}
}

func (c *Cookies) AddCookie(cookie *Cookie) {
	c.cookies[cookie.Name] = cookie
}

func (c *Cookies) GetCookie(name string) *Cookie {
	return c.cookies[name]
}

type Cookie struct {
	Name  string
	Value string
}

func InitCookie(name string, value string) *Cookie {
	return &Cookie{
		Name:  name,
		Value: value,
	}
}
