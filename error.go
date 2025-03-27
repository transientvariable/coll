package collection

const (
	ErrNoMoreElements   = collectionError("no more elements")
	ErrBoundsOutOfRange = collectionError("index bounds out of range")
	ErrCollectionEmpty  = collectionError("collection is empty")
	ErrNotFound         = collectionError("entry not found")
	ErrValueRequired    = collectionError("value is required")
)

type collectionError string

func (e collectionError) Error() string {
	return string(e)
}
