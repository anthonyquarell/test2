package constant

const (
	ServiceName = "electronic_product"

	MaxPageSize = 1000
)

// Key status
const (
	KeyStatusNew       = "new"
	KeyStatusActivated = "activated"
)

const (
	ProviderKasperksy = "kaspersky"
	ProviderMicrosoft = "microsoft"
)

var AllowedProviders = []string{
	ProviderKasperksy,
	ProviderMicrosoft,
}
