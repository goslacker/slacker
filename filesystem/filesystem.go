package filesystem

import "fmt"

type options struct {
	storage string
}

func prepareOptions(opts ...func(*options)) options {
	opt := options{
		storage: def,
	}
	for _, o := range opts {
		o(&opt)
	}
	return opt
}

func Put(path string, content []byte, opts ...func(*options)) error {
	opt := prepareOptions(opts...)
	return storages[opt.storage].Put(path, content)
}

func Get(path string, opts ...func(*options)) ([]byte, error) {
	opt := prepareOptions(opts...)
	return storages[opt.storage].Get(path)
}

func Del(path string, opts ...func(*options)) error {
	opt := prepareOptions(opts...)
	return storages[opt.storage].Del(path)
}

func Url(path string, opts ...func(*options)) (url string, err error) {
	opt := prepareOptions(opts...)
	storage := storages[opt.storage]
	if s, ok := storage.(CanGetUrl); ok {
		return s.Url(path)
	} else {
		err = fmt.Errorf("storage can not get url of file")
		return
	}
}
