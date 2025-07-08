package errs

type Err string

func (e Err) Error() string {
	return string(e)
}

// common errors
const (
	NoRows            = Err("err_no_rows")
	ServiceNA         = Err("service_not_available")
	NotAuthorized     = Err("not_authorized")
	ObjectNotFound    = Err("object_not_found")
	IncorrectPageSize = Err("incorrect_page_size")
	AlreadyExists     = Err("already_exists")

	EmptyData             = Err("empty_data")
	ProviderIDRequired    = Err("provider_id_required")
	ProductIDRequired     = Err("product_id_required")
	IDRequired            = Err("id_required")
	OrderIDRequired       = Err("order_id_required")
	CustomerPhoneRequired = Err("customer_phone_required")
	InvalidProviderID     = Err("invalid_provider_id")
	ValueRequired         = Err("value_required")
	InvalidPhone          = Err("invalid_phone")
	AlreadyCancelled      = Err("already_cancelled")
	AlreadyActivated      = Err("already_activated")
)

const (
	MethodNotSupported = Err("method_not_supported")
)

// ErrFull
type ErrFull struct {
	Err    error
	Desc   string
	Fields map[string]string
}

func (e ErrFull) Error() string {
	return e.Err.Error() + ", desc: " + e.Desc
}
