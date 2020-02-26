package libhttp

// Get
func Get(url string) *HttpRequest {
	return NewRequest(url, "GET")
}

// Post
func Post(url string) *HttpRequest {
	return NewRequest(url, "POST")
}

// Put
func Put(url string) *HttpRequest {
	return NewRequest(url, "PUT")
}

// Delete
func Delete(url string) *HttpRequest {
	return NewRequest(url, "DELETE")
}

// Head
func Head(url string) *HttpRequest {
	return NewRequest(url, "HEAD")
}