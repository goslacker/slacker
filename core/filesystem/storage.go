package filesystem

var storages = map[string]Storage{}
var def string

func SetDefaultStorage(name string) {
	def = name
}

func RegisterStorage(name string, storage Storage) {
	storages[name] = storage
}

func UnRegisterStorage(name string) {
	delete(storages, name)
}
