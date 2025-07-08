package model

import (
	"encoding/xml"
	"github.com/samber/lo"
	"regexp"

	"github.com/mechta-market/e-product/internal/service/provider/asbis/constant"
	providerModel "github.com/mechta-market/e-product/internal/service/provider/model"

	"time"
)

type OrderReq struct {
	XMLName                xml.Name    `xml:"SoftRequest"`
	InfoKind               string      `xml:"InfoKind"`
	ClientTransactionId    string      `xml:"ClientTransactionId"`
	RefClientTransactionId string      `xml:"RefClientTransactionId,omitempty"`
	TransactionType        string      `xml:"TransactionType"`
	WerkCode               string      `xml:"WerkCode"`
	TermNumber             string      `xml:"TermNumber"`
	TermDateTime           string      `xml:"TermDateTime"`
	SlipWidth              string      `xml:"SlipWidth"`
	ProductList            ProductList `xml:"ProductList"`
}

type ProductList struct {
	ProductItems []ProductItem `xml:"ProductItem"`
}

type ProductItem struct {
	ProductNumber string `xml:"ProductNumber"`
}

type OrderRep struct {
	XMLName             xml.Name             `xml:"SoftResponse"`
	AuthCode            string               `xml:"AuthCode"`
	ClientTransactionId string               `xml:"ClientTransactionId"`
	ErrorCode           string               `xml:"ErrorCode"`
	ErrorText           string               `xml:"ErrorText"`
	TransactionType     string               `xml:"TransactionType,omitempty"`
	ProductList         *ResponseProductList `xml:"ProductList,omitempty"`
}

type ResponseProductList struct {
	ProductItems []ResponseProductItem `xml:"ProductItem"`
}

type ResponseProductItem struct {
	ProductNumber string `xml:"ProductNumber"`
	Infos         *Infos `xml:"Infos,omitempty"`
	Slip          *Slip  `xml:"Slip,omitempty"`
}

type Infos struct {
	Info Info `xml:"Info"`
}

type Info struct {
	Article     string `xml:"Article"`
	Description string `xml:"Description"`
	Token       string `xml:"Token"`
	TokenLabel  string `xml:"TokenLabel"`
	UrlLabel    string `xml:"UrlLabel"`
}

type Slip struct {
	Lines []Line `xml:"Line"`
}

type Line struct {
	Text string `xml:",chardata"`
}

func EncodeActivateRequest(req *providerModel.OrderRequest) *OrderReq {
	return &OrderReq{
		InfoKind:            constant.InfoKind,
		TransactionType:     constant.TransactionTypeSell,
		WerkCode:            constant.WerkCode,
		SlipWidth:           constant.SLipWidth,
		TermDateTime:        time.Now().Format("02.01.2006 15:04:05"),
		ClientTransactionId: providerModel.GenerateUUID(),
		TermNumber:          req.ProductID, // From MDM

		ProductList: ProductList{
			ProductItems: []ProductItem{
				{
					ProductNumber: req.ProviderProductID,
				},
			},
		},
	}
}

func EncodeCancelRequest(req *providerModel.CancelRequest, newTransactionId string) *OrderReq {
	return &OrderReq{
		TransactionType:        constant.TransactionTypeCancel,
		WerkCode:               constant.WerkCode,
		SlipWidth:              constant.SLipWidth,
		ClientTransactionId:    newTransactionId, // CancelID аннулирования заказа
		RefClientTransactionId: lo.FromPtr(req.CancelID),
		TermNumber:             lo.FromPtr(req.ProductID),
		TermDateTime:           time.Now().Format("02.01.2006 15:04:05"),
		ProductList: ProductList{
			ProductItems: []ProductItem{
				{
					ProductNumber: *req.ProviderProductID,
				},
			},
		},
	}
}

func DecodeActivateResponse(resp OrderRep) *providerModel.OrderResponse {
	result := &providerModel.OrderResponse{
		Success:       resp.ErrorCode == "00000",
		TransactionID: resp.ClientTransactionId,
	}

	if result.Success && resp.ProductList != nil && len(resp.ProductList.ProductItems) > 0 {
		item := resp.ProductList.ProductItems[0]
		if item.Infos != nil {
			result.Value = item.Infos.Info.Token
		}

		if item.Slip != nil {
			receipt := ""
			for _, line := range item.Slip.Lines {
				receipt += line.Text + "\n"
			}
			result.Link = getLink(receipt)
		}

	}

	return result
}

func DecodeCancelResponse(rep OrderRep) *providerModel.CancelResponse {
	return &providerModel.CancelResponse{
		Success:       rep.ErrorCode == "00000",
		TransactionID: lo.ToPtr(rep.ClientTransactionId),
	}
}

func getLink(fullText string) *string {
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	match := urlPattern.FindString(fullText)

	return &match
}
