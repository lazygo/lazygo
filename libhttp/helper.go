package libhttp

// Get
func Get(url string) *HttpClient {
	return NewClient(url, "GET")
}

// Post
func Post(url string) *HttpClient {
	return NewClient(url, "POST")
}

// Put
func Put(url string) *HttpClient {
	return NewClient(url, "PUT")
}

// Delete
func Delete(url string) *HttpClient {
	return NewClient(url, "DELETE")
}

// Head
func Head(url string) *HttpClient {
	return NewClient(url, "HEAD")
}
