package model

import (
	"github.com/mechta-market/e-product/internal/constant"
	"github.com/mechta-market/e-product/internal/domain/common/util"
	"github.com/samber/lo"
)

func (e *Key) ValidateProviders() bool {
	if !lo.Contains(constant.AllowedProviders, *e.ProviderID) {
		return false
	}

	return true
}

func (e *Key) NormalizeAndValidatePhone() bool {
	if e.CustomerPhone == nil {
		return false
	}

	if !util.NormalizeAndValidatePhone(e.CustomerPhone) {
		return false
	}

	return true
}

func (e *Key) PhonesStr() string {
	return *e.CustomerPhone
}
