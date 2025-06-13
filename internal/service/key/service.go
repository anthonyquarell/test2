package key

//
//import (
//	"context"
//	"golang.org/x/sync/errgroup"
//	"log/slog"
//	"sync"
//	"time"
//)
//
//type Service struct {
//	keyService KeyServiceI
//	parentCtx  context.Context
//	wg         *sync.WaitGroup
//}
//
//func New(keyService KeyServiceI) *Service {
//	return &Service{
//		keyService: keyService,
//	}
//}
//
//func (s *Service) Start(ctx context.Context) {
//	if s.parentCtx != nil {
//		return
//	}
//	s.parentCtx = ctx
//	go s.service()
//}
//
//func (s *Service) Wait() {
//	s.wg.Wait()
//}
//
//func (s *Service) service() {
//	const delay = 15 * time.Second
//
//	time.Sleep(2 * time.Second)
//
//	for {
//		if s.parentCtx.Err() != nil {
//			return
//		}
//		s.wg.Add(1)
//		s.Routine()
//		time.Sleep(delay)
//	}
//}
//
//func (s *Service) Routine() {
//	defer s.wg.Done()
//	defer func() {
//		if r := recover(); r != nil {
//			slog.Error("key_service: Routine: panic recovery", "error", r)
//		}
//	}()
//
//	trueValue := true
//
//	ordList, _, err := s.ordService.List(context.Background(), &ordModel.ListReq{
//		WaitingAnyReceipt: &trueValue,
//	})
//	if err != nil {
//		slog.Error("key_service: Routine: ordService.List", "error", err)
//		return
//	}
//
//	if len(ordList) == 0 {
//		return
//	}
//
//	groupedOrdIds := s.groupOrdIdsByDb(ordList)
//
//	eg, egCtx := errgroup.WithContext(s.parentCtx)
//
//	for _, ids := range groupedOrdIds {
//		eg.Go(func() error {
//			if egCtx.Err() != nil {
//				return nil
//			}
//			s.serveDb(egCtx, ids)
//			return nil
//		})
//	}
//
//	if err = eg.Wait(); err != nil {
//		slog.Error("ord_wait_receipt_service: Routine: eg.Wait", "error", err)
//	}
//}
//
//func (s *Service) serveDb(ctx context.Context, ordIds []string) {
//	defer func() {
//		if r := recover(); r != nil {
//			slog.Error("key_service: serveDb: panic recovery", "error", r)
//		}
//	}()
//
//	const workerCount = 3
//
//	jobCh := make(chan string)
//
//	eg, egCtx := errgroup.WithContext(ctx)
//
//	for range workerCount {
//		eg.Go(func() error {
//			for ordId := range jobCh {
//				if egCtx.Err() != nil {
//					continue
//				}
//				s.serveOrd(egCtx, ordId)
//			}
//			return nil
//		})
//	}
//
//	go func() {
//		defer close(jobCh)
//
//		for _, ordId := range ordIds {
//			select {
//			case <-egCtx.Done():
//				return
//			default:
//				select {
//				case jobCh <- ordId:
//				case <-egCtx.Done():
//					return
//				}
//			}
//		}
//	}()
//
//	if err := eg.Wait(); err != nil {
//		slog.Error("key_service: serveDb: eg.Wait", "error", err)
//	}
//}
//
//func (s *Service) serveOrd(ctx context.Context, ordId string) {
//	defer func() {
//		if r := recover(); r != nil {
//			slog.Error("ord_wait_receipt_service: serveOrd: panic recovery", "error", r)
//		}
//	}()
//
//	falseValue := false
//
//	ord, _, err := s.ordService.Get(ctx, ordId, true)
//	if err != nil {
//		slog.Error("ord_wait_receipt_service: serveOrd: ordService.Get", "error", err)
//		return
//	}
//
//	if ord.CreatedAt.Before(time.Now().AddDate(0, 0, -1)) {
//		slog.Error("ord_wait_receipt_service: serveOrd: ord is too old", "ord_id", ord.Id, "created_at", ord.CreatedAt)
//		err = s.ordService.Set(ctx, &ordModel.Edit{
//			Id:                     ord.Id,
//			WaitingOrderReceipt:    &falseValue,
//			WaitingDeliveryReceipt: &falseValue,
//		})
//		if err != nil {
//			slog.Error("ord_wait_receipt_service: serveOrd: ordService.Set", "error", err)
//			return
//		}
//		return
//	}
//
//	editObj := &ordModel.Edit{
//		Id:      ord.Id,
//		Receipt: &ordModel.ReceiptEdit{},
//	}
//	changed := false
//
//	//slog.Info("ord_wait_receipt_service: serveOrd", "ord_id", ord.Id, "waiting_order_receipt", ord.WaitingOrderReceipt, "waiting_delivery_receipt", ord.WaitingDeliveryReceipt)
//
//	if ord.WaitingOrderReceipt {
//		receipt, err := s.shopService.GetReceipt(ctx, ord.DbName, ord.Id)
//		if err != nil {
//			slog.Error("ord_wait_receipt_service: serveOrd: shopService.GetReceipt", "error", err)
//			return
//		}
//
//		if receipt.Url != "" {
//			editObj.Receipt.URL = &receipt.Url
//			editObj.WaitingOrderReceipt = &falseValue
//			changed = true
//		}
//	}
//
//	if ord.WaitingDeliveryReceipt {
//		receipt, err := s.shopService.GetDeliveryReceipt(ctx, ord.DbName, ord.Id)
//		if err != nil {
//			slog.Error("ord_wait_receipt_service: serveOrd: shopService.GetDeliveryReceipt", "error", err)
//			return
//		}
//
//		if receipt.Url != "" {
//			editObj.Receipt.URLDelivery = &receipt.Url
//			editObj.WaitingDeliveryReceipt = &falseValue
//			changed = true
//		}
//	}
//
//	if changed {
//		err = s.ordService.Set(ctx, editObj)
//		if err != nil {
//			slog.Error("ord_wait_receipt_service: serveOrd: ordService.Set", "error", err)
//			return
//		}
//	}
//}
//
//func (s *Service) groupOrdIdsByDb(ords []*ordModel.Main) map[string][]string {
//	ordMap := make(map[string][]string, len(ords))
//
//	for _, ord := range ords {
//		ordMap[ord.DbName] = append(ordMap[ord.DbName], ord.Id)
//	}
//
//	return ordMap
//}
