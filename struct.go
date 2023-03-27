package jsonpath

type Map interface {
	Keys(keys ...string) Iterator
	Values(keys ...string) Iterator
	Set(key string, value any)
	Delete(key string)
}

type Array interface {
	Len() int
	Values(reverse bool, indexes ...int) Iterator
	Set(index int, value any)
}
