package url

type SchemeType string

const (
	HTTP  SchemeType = "http"
	HTTPS SchemeType = "https"
	FILE  SchemeType = "file"
	DATA  SchemeType = "data"
)

type URL struct {
	Host   string
	Path   string
	Scheme SchemeType
	Port   int
}

func (schema SchemeType) GetDefaultPort() int {
	switch schema {
	case HTTP:
		return 80
	case HTTPS:
		return 443
	case FILE:
		return 0
	case DATA:
		return 0
	default:
		return 80
	}
}
