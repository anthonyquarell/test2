package comportal

type SendReq struct {
	Method string
	DbName string
	Path   string
	ReqObj any
	RepObj any
	Params map[string]string
}
