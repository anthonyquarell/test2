package key

import (
	"context"
	"errors"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"

	"github.com/mechta-market/e-product/internal/constant"
	commonModel "github.com/mechta-market/e-product/internal/domain/common/model"
	"github.com/mechta-market/e-product/internal/domain/key/model"
	"github.com/mechta-market/e-product/internal/errs"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"
	"github.com/mechta-market/e-product/internal/usecase/key/mocks"
)

// TODO: refactor Load & Activate tests

type usecaseTest struct {
	service         *mocks.KeyServiceI
	mdmService      *mocks.MdmServiceI
	providerService *mocks.ProviderServiceI
	providers       map[string]ProviderServiceI
	usecase         *Usecase
}

func newTest() *usecaseTest {
	service := new(mocks.KeyServiceI)
	mdmSerivce := new(mocks.MdmServiceI)
	providerService := new(mocks.ProviderServiceI)

	providers := map[string]ProviderServiceI{
		"provider-1": providerService,
	}

	return &usecaseTest{
		service:         service,
		mdmService:      mdmSerivce,
		providerService: providerService,
		providers:       providers,
	}
}

func TestUsecase_List(t *testing.T) {
	type args struct {
		pageSize int64
	}
	tests := []struct {
		name        string
		args        args
		setupMock   func(t *usecaseTest, req *model.ListReq)
		wantErr     bool
		wantItems   int
		wantTotal   int64
		expectedErr error
	}{
		{
			name: "success",
			args: args{pageSize: 10},
			setupMock: func(ut *usecaseTest, req *model.ListReq) {
				ut.service.On("List", mock.Anything, req).Return(
					[]*model.Main{
						{ID: "key-1"},
						{ID: "key-2"},
					}, int64(2), nil,
				).Once()
			},
			wantErr:   false,
			wantItems: 2,
			wantTotal: 2,
		},
		{
			name: "invalid page size",
			args: args{pageSize: constant.MaxPageSize + 1},
			setupMock: func(ut *usecaseTest, req *model.ListReq) {
			},
			wantErr:     true,
			expectedErr: errs.IncorrectPageSize,
		},
		{
			name: "service.List returns error",
			args: args{pageSize: 5},
			setupMock: func(ut *usecaseTest, req *model.ListReq) {
				ut.service.On("List", mock.Anything, req).Return(nil, int64(0), errors.New("some error")).Once()
			},
			wantErr:     true,
			expectedErr: errors.New("some error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := newTest()
			ut.usecase = New(ut.service, ut.mdmService, ut.providers)

			req := &model.ListReq{
				ListParams: commonModel.ListParams{
					PageSize: tt.args.pageSize,
				},
			}

			if tt.setupMock != nil {
				tt.setupMock(ut, req)
			}

			items, total, err := ut.usecase.List(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorContains(t, err, tt.expectedErr.Error())
				}
				return
			}

			assert.NoError(t, err)
			assert.Len(t, items, tt.wantItems)
			assert.Equal(t, tt.wantTotal, total)

			ut.service.AssertExpectations(t)
		})
	}
}

//func TestUsecase_Load(t *testing.T) {
//	tests := []struct {
//		name        string
//		input       []*model.Edit
//		setupMock   func(ut *usecaseTest, input []*model.Edit)
//		expectedErr error
//	}{
//		{
//			name: "successful load",
//			input: []*model.Edit{
//				{ID: lo.ToPtr("1")},
//				{ID: lo.ToPtr("2")},
//			},
//			setupMock: func(ut *usecaseTest, input []*model.Edit) {
//				ut.service.On("Load", mock.Anything, input).Return(nil).Once()
//			},
//			expectedErr: nil,
//		},
//		{
//			name: "service.Load returns error",
//			input: []*model.Edit{
//				{ID: lo.ToPtr("fail")},
//			},
//			setupMock: func(ut *usecaseTest, input []*model.Edit) {
//				ut.service.On("Load", mock.Anything, input).Return(errors.New("load failed")).Once()
//			},
//			expectedErr: errors.New("load failed"),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ut := newTest()
//			ut.usecase = New(ut.service, ut.mdmService, ut.providers)
//
//			if tt.setupMock != nil {
//				tt.setupMock(ut, tt.input)
//			}
//
//			err := ut.usecase.Load(context.Background(), tt.input)
//
//			if tt.expectedErr != nil {
//				assert.Error(t, err)
//				assert.ErrorContains(t, err, tt.expectedErr.Error())
//			} else {
//				assert.NoError(t, err)
//			}
//
//			ut.service.AssertExpectations(t)
//		})
//	}
//}

func TestUsecase_Get(t *testing.T) {
	tests := []struct {
		name        string
		keyID       string
		setupMock   func(t *usecaseTest, id string)
		expectedKey *model.Main
		wantErr     bool
	}{
		{
			name:  "success",
			keyID: "123",
			setupMock: func(ut *usecaseTest, id string) {
				mainKey := &model.Main{ID: id}
				ut.service.On("Get", mock.Anything, id, true).Return(mainKey, true, nil).Once()
			},
			expectedKey: &model.Main{ID: "123"},
			wantErr:     false,
		},
		{
			name: "not found",
			setupMock: func(ut *usecaseTest, id string) {
				ut.service.On("Get", mock.Anything, id, true).Return(nil, false, nil).Once()
			},
			expectedKey: nil,
			wantErr:     false,
		},
		{
			name:  "service.Get returns error",
			keyID: "failed",
			setupMock: func(ut *usecaseTest, id string) {
				ut.service.On("Get", mock.Anything, id, true).Return(nil, false, errors.New("some error")).Once()
			},
			expectedKey: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := newTest()
			ut.usecase = New(ut.service, ut.mdmService, ut.providers)

			if tt.setupMock != nil {
				tt.setupMock(ut, tt.keyID)
			}

			result, err := ut.usecase.Get(context.Background(), tt.keyID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedKey, result)

			ut.service.AssertExpectations(t)
		})
	}
}

func TestUsecase_GetCatalog(t *testing.T) {
	tests := []struct {
		name            string
		providerID      string
		setupMock       func(ut *usecaseTest, providerID string)
		expectedCatalog []*providerModel.CatalogResponse
		expectedErr     error
	}{
		{
			name:       "success",
			providerID: "provider-1",
			setupMock: func(ut *usecaseTest, providerID string) {
				catalog := []*providerModel.CatalogResponse{
					{ProviderProductID: lo.ToPtr("prod-1"), Name: lo.ToPtr("Product 1")},
					{ProviderProductID: lo.ToPtr("prod-2"), Name: lo.ToPtr("Product 2")},
				}
				ut.providerService.On("ListCatalog", mock.Anything, providerID).Return(catalog, nil).Once()
			},
			expectedCatalog: []*providerModel.CatalogResponse{
				{ProviderProductID: lo.ToPtr("prod-1"), Name: lo.ToPtr("Product 1")},
				{ProviderProductID: lo.ToPtr("prod-2"), Name: lo.ToPtr("Product 2")},
			},
			expectedErr: nil,
		},
		{
			name:            "provider not found",
			providerID:      "unknown-provider",
			setupMock:       func(ut *usecaseTest, providerID string) {},
			expectedCatalog: nil,
			expectedErr:     errors.New("providerService.GetProvider"),
		},
		{
			name:       "provider service error",
			providerID: "provider-1",
			setupMock: func(ut *usecaseTest, providerID string) {
				ut.providerService.On("ListCatalog", mock.Anything, providerID).Return(nil, errors.New("api error")).Once()
			},
			expectedCatalog: nil,
			expectedErr:     errors.New("api error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := newTest()
			ut.usecase = New(ut.service, ut.mdmService, ut.providers)

			if tt.setupMock != nil {
				tt.setupMock(ut, tt.providerID)
			}

			result, err := ut.usecase.GetCatalog(context.Background(), tt.providerID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCatalog, result)
			}

			if tt.providerID == "provider-1" {
				ut.providerService.AssertExpectations(t)
			}
		})
	}
}

//func TestUsecase_Activate(t *testing.T) {
//	tests := []struct {
//		name          string
//		productID     string
//		orderID       string
//		customerPhone string
//		setupMock     func(ut *usecaseTest)
//		expectedKey   *model.Main
//		expectedErr   error
//	}{
//		{
//			name:          "success - order activated via provider",
//			productID:     "prod-1",
//			orderID:       "ord-1",
//			customerPhone: "+77001112233",
//			setupMock: func(ut *usecaseTest) {
//				mdmProduct := &mdmModel.Product{
//					ProviderID:         "provider-1",
//					ProductID:          "prod-1",
//					ProviderProductID:  "prov-prod-1",
//					ProviderExternalID: lo.ToPtr("ext-1"),
//					PromotionKey:       lo.ToPtr("promo-1"),
//				}
//				ut.mdmService.On("FindProduct", mock.Anything, lo.ToPtr("prod-1")).Return(mdmProduct, true, nil).Once()
//
//				providerResp := &providerModel.OrderResponse{
//					//TransactionID: lo.ToPtr("key-123"),
//					OrderID: lo.ToPtr("prov-ord-1"),
//					Value:   "secret-value",
//				}
//				ut.providerService.On("CreateOrder", mock.Anything, mock.MatchedBy(func(req *providerModel.OrderRequest) bool {
//					return req.ProductID == "prod-1" && req.OrderID == "ord-1" && req.CustomerPhone == "+77001112233"
//				})).Return(providerResp, nil).Once()
//				ut.service.On(
//					"GetByOrderID",
//					mock.Anything,
//					"ord-1",
//					false,
//				).Return(nil, false, nil).Once()
//				ut.service.On(
//					"ActivateWithProvider",
//					mock.Anything,
//					mock.MatchedBy(func(req *providerModel.OrderRequest) bool {
//						return req.ProductID == "prod-1" && req.OrderID == "ord-1"
//					}),
//					providerResp,
//				).Return(nil).Once()
//			},
//			expectedKey: &model.Main{
//				Value: "secret-value",
//			},
//			expectedErr: nil,
//		},
//		{
//			name:          "error - MDM product not found",
//			productID:     "unknown-prod",
//			orderID:       "ord-2",
//			customerPhone: "+77001112233",
//			setupMock: func(ut *usecaseTest) {
//				ut.mdmService.On("FindProduct", mock.Anything, lo.ToPtr("unknown-prod")).Return(nil, false, errors.New("продукт не найден в MDM")).Once()
//				ut.service.On(
//					"GetByOrderID",
//					mock.Anything,
//					"ord-2",
//					false,
//				).Return(nil, false, nil).Once()
//			},
//			expectedKey: nil,
//			expectedErr: errors.New("продукт не найден в MDM"),
//		},
//		{
//			name:          "error - provider service fails",
//			productID:     "prod-2",
//			orderID:       "ord-3",
//			customerPhone: "+77001112233",
//			setupMock: func(ut *usecaseTest) {
//				mdmProduct := &mdmModel.Product{
//					ProviderID:        "provider-1",
//					ProductID:         "prod-2",
//					ProviderProductID: "prov-prod-2",
//				}
//				ut.service.On(
//					"GetByOrderID",
//					mock.Anything,
//					"ord-3",
//					false,
//				).Return(nil, false, nil).Once()
//				ut.mdmService.On("FindProduct", mock.Anything, lo.ToPtr("prod-2")).Return(mdmProduct, true, nil).Once()
//
//				ut.providerService.On("CreateOrder", mock.Anything, mock.Anything).Return(nil, errors.New("provider service down")).Once()
//
//				ut.service.On("ActivateWithPool", mock.Anything, mock.MatchedBy(func(req *providerModel.OrderRequest) bool {
//					return req.ProductID == "prod-2" && req.OrderID == "ord-3"
//				})).Return(nil, errors.New("no available keys in pool")).Once()
//			},
//			expectedKey: nil,
//			expectedErr: errors.New("serivce.ActivateWithPool: no available keys in pool"),
//		},
//		{
//			name:          "error - provider not configured",
//			productID:     "prod-3",
//			orderID:       "ord-4",
//			customerPhone: "+77001112233",
//			setupMock: func(ut *usecaseTest) {
//				ut.service.On(
//					"GetByOrderID",
//					mock.Anything,
//					"ord-4",
//					false,
//				).Return(nil, false, nil).Once()
//				mdmProduct := &mdmModel.Product{
//					ProviderID:        "unknown-provider",
//					ProductID:         "prod-3",
//					ProviderProductID: "prov-prod-3",
//				}
//				ut.mdmService.On("FindProduct", mock.Anything, lo.ToPtr("prod-3")).Return(mdmProduct, true, nil).Once()
//			},
//			expectedKey: nil,
//			expectedErr: errors.New("Услуги провайдера не подключены"),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			ut := newTest()
//			ut.usecase = New(ut.service, ut.mdmService, ut.providers)
//
//			if tt.setupMock != nil {
//				tt.setupMock(ut)
//			}
//
//			result, err := ut.usecase.Activate(context.Background(), tt.productID, tt.orderID, tt.customerPhone)
//
//			if tt.expectedErr != nil {
//				assert.Error(t, err)
//				assert.ErrorContains(t, err, tt.expectedErr.Error())
//				assert.Nil(t, result)
//			} else {
//				assert.NoError(t, err)
//				if tt.expectedKey != nil && result != nil {
//					assert.Equal(t, tt.expectedKey.Value, result.Value)
//				}
//			}
//
//			ut.service.AssertExpectations(t)
//			ut.mdmService.AssertExpectations(t)
//			ut.providerService.AssertExpectations(t)
//		})
//	}
//}

func TestUsecase_Cancel(t *testing.T) {
	tests := []struct {
		name        string
		orderID     string
		setupMock   func(ut *usecaseTest)
		expectedID  *string
		expectedErr error
	}{
		{
			name:    "success",
			orderID: "ord-1",
			setupMock: func(ut *usecaseTest) {
				main := &model.Main{
					ID:                "key-1",
					ProviderID:        "provider-1",
					ProductID:         "prod-1",
					ProviderProductID: "prov-prod-1",
					CustomerPhone:     "+77001112233",
				}
				ut.service.On("GetByOrderID", mock.Anything, "ord-1", true).Return(main, true, nil).Once()

				ut.providerService.On("CancelOrder", mock.Anything, mock.Anything).Return(&providerModel.CancelResponse{Success: true}, nil).Once()

				ut.service.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedID:  lo.ToPtr("key-1"),
			expectedErr: nil,
		},
		{
			name:        "validation error - empty order CancelID",
			orderID:     "",
			setupMock:   func(ut *usecaseTest) {},
			expectedID:  nil,
			expectedErr: errs.OrderIDRequired,
		},
		{
			name:    "order not found",
			orderID: "non-existent",
			setupMock: func(ut *usecaseTest) {
				ut.service.On("GetByOrderID", mock.Anything, "non-existent", true).Return(nil, false, errs.ObjectNotFound).Once()
			},
			expectedID:  nil,
			expectedErr: errs.ObjectNotFound,
		},
		{
			name:    "provider not found",
			orderID: "ord-1",
			setupMock: func(ut *usecaseTest) {
				main := &model.Main{
					ID:         "key-1",
					ProviderID: "unknown-provider",
				}
				ut.service.On("GetByOrderID", mock.Anything, "ord-1", true).Return(main, true, nil).Once()
			},
			expectedID:  nil,
			expectedErr: errors.New("Услуги провайдера не подключены"),
		},
		{
			name:    "provider service error",
			orderID: "ord-1",
			setupMock: func(ut *usecaseTest) {
				main := &model.Main{
					ID:         "key-1",
					ProviderID: "provider-1",
				}
				ut.service.On("GetByOrderID", mock.Anything, "ord-1", true).Return(main, true, nil).Once()
				ut.providerService.On("CancelOrder", mock.Anything, mock.Anything).Return(nil, errors.New("provider error")).Once()
			},
			expectedID:  nil,
			expectedErr: errors.New("provider error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ut := newTest()
			ut.usecase = New(ut.service, ut.mdmService, ut.providers)

			if tt.setupMock != nil {
				tt.setupMock(ut)
			}

			result, err := ut.usecase.Cancel(context.Background(), tt.orderID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, result)
			}

			ut.service.AssertExpectations(t)
			if tt.orderID != "" && tt.orderID != "non-existent" {
				ut.mdmService.AssertExpectations(t)
				if tt.name != "provider not found" {
					ut.providerService.AssertExpectations(t)
				}
			}
		})
	}
}

func TestUsecase_validateActivate(t *testing.T) {
	tests := []struct {
		productID     string
		orderID       string
		customerPhone string
		expectedErr   error
	}{
		{
			productID:     "prod-1",
			orderID:       "ord-1",
			customerPhone: "+77001112233",
			expectedErr:   nil,
		},
		{
			productID:     "  prod-1  ",
			orderID:       "  ord-1  ",
			customerPhone: "  +77001112233  ",
			expectedErr:   nil,
		},
		{
			productID:     "",
			orderID:       "ord-1",
			customerPhone: "+77001112233",
			expectedErr:   errs.ProductIDRequired,
		},
		{
			productID:     "prod-1",
			orderID:       "",
			customerPhone: "+77001112233",
			expectedErr:   errs.OrderIDRequired,
		},
		{
			productID:     "prod-1",
			orderID:       "  ",
			customerPhone: "+77001112233",
			expectedErr:   errs.OrderIDRequired,
		},
		{
			productID:     "prod-1",
			orderID:       "ord-1",
			customerPhone: "",
			expectedErr:   errs.CustomerPhoneRequired,
		},
		{
			productID:     "prod-1",
			orderID:       "ord-1",
			customerPhone: "+7700",
			expectedErr:   errors.New("Номер телефона клиента не прошел валидацию"),
		},
		{
			productID:     "prod-1",
			orderID:       "ord-1",
			customerPhone: "+7700abc1234",
			expectedErr:   errors.New("Номер телефона клиента не прошел валидацию"),
		},
		{
			productID:     "",
			orderID:       "",
			customerPhone: "",
			expectedErr:   errs.OrderIDRequired,
		},
		{
			productID:     "prod-1",
			orderID:       "ord-1",
			customerPhone: "+77001112233",
			expectedErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			ut := newTest()
			ut.usecase = New(ut.service, ut.mdmService, ut.providers)

			ut.service.On("GetByOrderID", mock.Anything, strings.TrimSpace(tt.orderID), false).Return(nil, false, nil).Once()

			err := ut.usecase.validateActivate(context.Background(), tt.orderID, tt.productID, tt.customerPhone)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
