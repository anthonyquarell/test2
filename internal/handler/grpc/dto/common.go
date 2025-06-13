package dto

import (
	commonModel "github.com/mechta-market/e-product/internal/domain/common/model"
	"github.com/mechta-market/e-product/pkg/proto/common"
)

func DecodeListParams(listParams *common.ListParamsSt) commonModel.ListParams {
	if listParams == nil {
		return commonModel.ListParams{}
	}

	return commonModel.ListParams{
		Page:           listParams.Page,
		PageSize:       listParams.PageSize,
		WithTotalCount: listParams.WithTotalCount,
		OnlyCount:      listParams.OnlyCount,
		SortName:       listParams.SortName,
		Sort:           listParams.Sort,
	}
}
