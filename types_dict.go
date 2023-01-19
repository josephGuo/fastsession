package fastsession

//go:generate msgp

// Dict memory store.
type Dict struct {
	KV map[string]interface{}
}
