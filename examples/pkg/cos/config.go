package cos

type Config struct {
	BucketURL string `json:"bucket_url" toml:"bucket_url"`
	SecretID  string `json:"secret_id" toml:"secret_id"`
	SecretKey string `json:"secret_key" toml:"secret_key"`
}

/**
rawurl := "https://img-domain-1303896251.cos.ap-nanjing.myqcloud.com"
Id := "NKIDXYIn113YFgtqXypfsqX02rRkwud5AJ8j"
Key := "Yw50e7MQIuQkpgRyKuDUYOxFLOCnFTWJ"
*/
