package util

import (
	"regexp"
	"strings"

	"github.com/mechta-market/e-product/internal/domain/common/model"
	"github.com/mechta-market/e-product/internal/errs"
)

const (
	defaultMaxPageSize int64 = 100
)

var (
	phoneRegexp = regexp.MustCompile(`^[1-9][0-9]{10,30}$`)
)

func RequirePageSize(pars model.ListParams, maxPageSize int64) error {
	if maxPageSize == 0 {
		maxPageSize = defaultMaxPageSize
	}

	if pars.PageSize == 0 || pars.PageSize > maxPageSize {
		return errs.IncorrectPageSize
	}

	return nil
}

func NormalizeAndValidatePhone(phone *string) bool {
	if phone == nil {
		return false
	}
	l := len(*phone)
	if l > 1 {
		if (*phone)[0] == '+' {
			*phone = (*phone)[1:]
			l--
		}
		if l == 10 && (*phone)[0] == '7' {
			*phone = "7" + *phone
		} else if l == 11 && strings.HasPrefix(*phone, "87") {
			*phone = "7" + (*phone)[1:]
		}
	}
	return phoneRegexp.MatchString(*phone)
}
