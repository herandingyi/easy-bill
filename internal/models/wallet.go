package models

import (
	"math/big"
	"time"
)

type Wallet struct {
	Id                 int64
	UserId             int64
	Type               int
	AccountNumerator   int64
	AccountDenominator int64
}

type WalletLog struct {
	Id                int64
	CommandId         int64
	UserId            int64
	Type              int
	BeforeNumerator   int64
	BeforeDenominator int64
	Inc               float64
	AfterNumerator    int64
	AfterDenominator  int64
	CreateTime        time.Time
}

func (w *Wallet) Inc(numerator, denominator int64) (int64, int64) {
	a := big.NewRat(w.AccountNumerator, w.AccountDenominator)
	b := big.NewRat(numerator, denominator)
	c := new(big.Rat).Add(a, b)
	return c.Num().Int64(), c.Denom().Int64()
}
