package model

import (
	"encoding/xml"
	"github.com/mechta-market/e-product/internal/service/provider/asbis/constant"
	serviceModel "github.com/mechta-market/e-product/internal/service/provider/asbis/model"
	"regexp"

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

func EncodeActivateRequest(req *serviceModel.OrderReq, transactionId string) *OrderReq {
	return &OrderReq{
		InfoKind:            constant.InfoKind,
		TransactionType:     constant.TransactionTypeSell,
		WerkCode:            constant.WerkCode, // From documentation
		SlipWidth:           constant.SLipWidth,
		TermDateTime:        time.Now().Format("02.01.2006 15:04:05"),
		ClientTransactionId: transactionId,
		TermNumber:          *req.TermNumber, // From MDM

		ProductList: ProductList{
			ProductItems: []ProductItem{
				{
					ProductNumber: *req.ProductNumber,
				},
			},
		},
	}
}

func EncodeCancelRequest(req *serviceModel.CancelReq, newTransactionId string) *OrderReq {
	return &OrderReq{
		TransactionType:        constant.TransactionTypeCancel,
		WerkCode:               constant.WerkCode,
		SlipWidth:              constant.SLipWidth,
		ClientTransactionId:    newTransactionId, // ID аннулирования заказа
		RefClientTransactionId: *req.OriginalTransactionID,
		TermNumber:             *req.TermNumber,
		TermDateTime:           time.Now().Format("02.01.2006 15:04:05"),
		ProductList: ProductList{
			ProductItems: []ProductItem{
				{
					ProductNumber: *req.ProductNumber,
				},
			},
		},
	}
}

func DecodeActivateResponse(resp OrderRep) *serviceModel.OrderRep {
	result := &serviceModel.OrderRep{
		Success:  resp.ErrorCode == "00000",
		ErrorMsg: resp.ErrorText,
		ID:       &resp.ClientTransactionId,
	}

	if result.Success && resp.ProductList != nil && len(resp.ProductList.ProductItems) > 0 {
		item := resp.ProductList.ProductItems[0]
		if item.Infos != nil {
			result.Value = &item.Infos.Info.Token
		}

		if item.Slip != nil {
			receipt := ""
			for _, line := range item.Slip.Lines {
				receipt += line.Text + "\n"
			}
			result.Receipt = &receipt
			result.Link = getLink(receipt)
		}

	}

	return result
}

func DecodeCancelResponse(rep OrderRep) *serviceModel.CancelRep {
	return &serviceModel.CancelRep{
		Success:               rep.ErrorCode == "00000",
		ErrorMessage:          rep.ErrorText,
		OriginalTransactionID: rep.ClientTransactionId,
	}
}

func getLink(fullText string) *string {
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	match := urlPattern.FindString(fullText)

	return &match
}
