package types

type Object[T any] map[string]T
type Array[T any] []T
type Dict = Object[any]
