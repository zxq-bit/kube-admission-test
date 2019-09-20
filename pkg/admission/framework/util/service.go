package util

type ServiceConfig struct {
	Name      string
	Namespace string
	RootPath  string
	CABundle  []byte
}
