package database

type DataError string

func (de DataError) Error() string {
	return string(de)
}
