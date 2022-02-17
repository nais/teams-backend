package apiserver

type ID struct {
	ID string `json:"id" path:"id" format:"uuid"`
}
