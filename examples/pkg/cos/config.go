package cos

type Config struct {
	BucketURL string `json:"bucket_url" toml:"bucket_url"`
	SecretID  string `json:"secret_id" toml:"secret_id"`
	SecretKey string `json:"secret_key" toml:"secret_key"`
}

/**
rawurl := "https://img-domain-00000.cos.ap-nanjing.myqcloud.com"
Id := "NKI....AJ8j"
Key := "Yw50e...FTWJ"
*/
