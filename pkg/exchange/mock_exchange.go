package exchange

type MockExchange struct {
	commissionPercent float64
}

func NewMockExchange(commissionPercent float64) *MockExchange {
	return &MockExchange{commissionPercent: commissionPercent}
}

func (e *MockExchange) Buy(orderRequest OrderRequest) (orderResult OrderResult, err error) {
	basePrice := orderRequest.Price * float64(orderRequest.Count)
	commission := basePrice * e.commissionPercent

	return OrderResult{
		Count:      orderRequest.Count,
		LotPrice:   basePrice + commission,
		Commission: commission,
	}, nil
}

func (e *MockExchange) Sell(orderRequest OrderRequest) (orderResult OrderResult, err error) {
	basePrice := orderRequest.Price * float64(orderRequest.Count)
	commission := basePrice * e.commissionPercent

	return OrderResult{
		Count:      orderRequest.Count,
		LotPrice:   basePrice - commission,
		Commission: commission,
	}, nil
}
