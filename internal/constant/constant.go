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

	ProviderComportal = "67fdc221-1211-11ea-a211-005056b6e0df"
)

var AllowedProviders = []string{
	ProviderKasperksy,
	ProviderMicrosoft,
}
