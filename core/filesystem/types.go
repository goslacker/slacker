package filesystem

type Storage interface {
	Put(path string, content []byte) error
	Del(path string) error
	Get(path string) ([]byte, error)
}

type CanGetUrl interface {
	Url(path string) (url string, err error)
}
