package decoder

import (
	"strings"
	"testing"

	binarycodec "github.com/Peersyst/xrpl-go/binary-codec"
	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
	"go.uber.org/zap"
)

// Test fixtures: valid XRPL classic addresses and 32-byte (64 hex char) IDs.
const (
	testAccount = "r2oU84CFuT4MgmrDejBaoyHNvovpMSPiA"
	testIssuer  = "r33hypJXDs47LVpmvta7hMW9pR8DYeBtkW"
	testOther   = "r3AthBf5eW4b9ujLoXNHFeeEJsK3PtJDea"

	testPubKey = "ED54C1E3427192B879EBD6F1FD7306058EAE6DAF7D95B2655B053885FE7722A712"
	testSig    = "A7C6C1EE9989F9F195A02BEA4DCFEBB887E4CA1F4D30083C84616E0FD1BCA4F4C1B84A6DA26A44B94FBBDA67FB603C78995361DEAF8120093959C639E9255702"
)

var (
	vaultID  = strings.Repeat("1", 64)
	brokerID = strings.Repeat("2", 64)
	loanID   = strings.Repeat("3", 64)
	domainID = strings.Repeat("4", 64)
)

// mapToProto encodes a flat transaction to the XRPL binary format, decodes it
// back through the production decoder, and maps it to protobuf — exercising the
// full wire -> proto pipeline that the fetcher relies on.
func mapToProto(t *testing.T, flat map[string]any) *pbxrpl.Transaction {
	t.Helper()

	blobHex, err := binarycodec.Encode(flat)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	d := NewDecoder(zap.NewNop())
	decoded, err := d.DecodeTransactionFromHex(blobHex)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	proto, err := d.mapper.MapTransactionToProto(decoded, nil, nil, []byte{0xAB}, 0, "tesSUCCESS")
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	return proto
}

func eqStr(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %q, want %q", field, got, want)
	}
}

func eqU32(t *testing.T, field string, got, want uint32) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", field, got, want)
	}
}

// xrpAmount builds an XRP (drops) amount, iouAmount an issued-currency amount.
func xrpAmount(drops string) any { return drops }
func iouAmount(currency, issuer, value string) any {
	return map[string]any{"currency": currency, "issuer": issuer, "value": value}
}

func TestMapVaultTransactions(t *testing.T) {
	tests := []struct {
		name   string
		flat   map[string]any
		verify func(*testing.T, *pbxrpl.Transaction)
	}{
		{
			name: "VaultCreate",
			flat: map[string]any{
				"TransactionType":  "VaultCreate",
				"Account":          testAccount,
				"Asset":            map[string]any{"currency": "USD", "issuer": testIssuer},
				"AssetsMaximum":    "1000000000",
				"MPTokenMetadata":  "DEADBEEF",
				"DomainID":         domainID,
				"WithdrawalPolicy": 1,
				"Data":             "CAFE",
				"Scale":            6,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				eqStr(t, "TxType", tx.TxType, "VaultCreate")
				eqStr(t, "Account", tx.Account, testAccount)
				v := tx.GetVaultCreate()
				if v == nil {
					t.Fatal("VaultCreate detail is nil")
				}
				if v.Asset == nil {
					t.Fatal("Asset is nil")
				}
				eqStr(t, "Asset.Currency", v.Asset.Currency, "USD")
				eqStr(t, "Asset.Issuer", v.Asset.Issuer, testIssuer)
				eqStr(t, "AssetsMaximum", v.AssetsMaximum, "1000000000")
				eqStr(t, "MptokenMetadata", v.MptokenMetadata, "DEADBEEF")
				eqStr(t, "DomainId", v.DomainId, domainID)
				eqU32(t, "WithdrawalPolicy", v.WithdrawalPolicy, 1)
				eqStr(t, "Data", v.Data, "CAFE")
				eqU32(t, "Scale", v.Scale, 6)
			},
		},
		{
			name: "VaultSet",
			flat: map[string]any{
				"TransactionType": "VaultSet",
				"Account":         testAccount,
				"VaultID":         vaultID,
				"AssetsMaximum":   "2000000000",
				"DomainID":        domainID,
				"Data":            "BEEF",
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetVaultSet()
				if v == nil {
					t.Fatal("VaultSet detail is nil")
				}
				eqStr(t, "VaultId", v.VaultId, vaultID)
				eqStr(t, "AssetsMaximum", v.AssetsMaximum, "2000000000")
				eqStr(t, "DomainId", v.DomainId, domainID)
				eqStr(t, "Data", v.Data, "BEEF")
			},
		},
		{
			name: "VaultDelete",
			flat: map[string]any{
				"TransactionType": "VaultDelete",
				"Account":         testAccount,
				"VaultID":         vaultID,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetVaultDelete()
				if v == nil {
					t.Fatal("VaultDelete detail is nil")
				}
				eqStr(t, "VaultId", v.VaultId, vaultID)
			},
		},
		{
			name: "VaultDeposit",
			flat: map[string]any{
				"TransactionType": "VaultDeposit",
				"Account":         testAccount,
				"VaultID":         vaultID,
				"Amount":          xrpAmount("1000000"),
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetVaultDeposit()
				if v == nil {
					t.Fatal("VaultDeposit detail is nil")
				}
				eqStr(t, "VaultId", v.VaultId, vaultID)
				if v.Amount == nil {
					t.Fatal("Amount is nil")
				}
				eqStr(t, "Amount.Value", v.Amount.Value, "1000000")
			},
		},
		{
			name: "VaultWithdraw",
			flat: map[string]any{
				"TransactionType": "VaultWithdraw",
				"Account":         testAccount,
				"VaultID":         vaultID,
				"Amount":          iouAmount("USD", testIssuer, "50"),
				"Destination":     testOther,
				"DestinationTag":  7,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetVaultWithdraw()
				if v == nil {
					t.Fatal("VaultWithdraw detail is nil")
				}
				eqStr(t, "VaultId", v.VaultId, vaultID)
				if v.Amount == nil {
					t.Fatal("Amount is nil")
				}
				eqStr(t, "Amount.Value", v.Amount.Value, "50")
				eqStr(t, "Amount.Currency", v.Amount.Currency, "USD")
				eqStr(t, "Amount.Issuer", v.Amount.Issuer, testIssuer)
				eqStr(t, "Destination", v.Destination, testOther)
				eqU32(t, "DestinationTag", v.DestinationTag, 7)
			},
		},
		{
			name: "VaultClawback",
			flat: map[string]any{
				"TransactionType": "VaultClawback",
				"Account":         testAccount,
				"VaultID":         vaultID,
				"Holder":          testOther,
				"Amount":          xrpAmount("250"),
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetVaultClawback()
				if v == nil {
					t.Fatal("VaultClawback detail is nil")
				}
				eqStr(t, "VaultId", v.VaultId, vaultID)
				eqStr(t, "Holder", v.Holder, testOther)
				if v.Amount == nil {
					t.Fatal("Amount is nil")
				}
				eqStr(t, "Amount.Value", v.Amount.Value, "250")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.verify(t, mapToProto(t, tc.flat))
		})
	}
}

func TestMapLendingTransactions(t *testing.T) {
	tests := []struct {
		name   string
		flat   map[string]any
		verify func(*testing.T, *pbxrpl.Transaction)
	}{
		{
			name: "LoanBrokerSet",
			flat: map[string]any{
				"TransactionType":      "LoanBrokerSet",
				"Account":              testAccount,
				"VaultID":              vaultID,
				"LoanBrokerID":         brokerID,
				"Data":                 "F00D",
				"ManagementFeeRate":    100,
				"DebtMaximum":          "5000",
				"CoverRateMinimum":     200,
				"CoverRateLiquidation": 300,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanBrokerSet()
				if v == nil {
					t.Fatal("LoanBrokerSet detail is nil")
				}
				eqStr(t, "VaultId", v.VaultId, vaultID)
				eqStr(t, "LoanBrokerId", v.LoanBrokerId, brokerID)
				eqStr(t, "Data", v.Data, "F00D")
				eqU32(t, "ManagementFeeRate", v.ManagementFeeRate, 100)
				eqStr(t, "DebtMaximum", v.DebtMaximum, "5000")
				eqU32(t, "CoverRateMinimum", v.CoverRateMinimum, 200)
				eqU32(t, "CoverRateLiquidation", v.CoverRateLiquidation, 300)
			},
		},
		{
			name: "LoanBrokerDelete",
			flat: map[string]any{
				"TransactionType": "LoanBrokerDelete",
				"Account":         testAccount,
				"LoanBrokerID":    brokerID,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanBrokerDelete()
				if v == nil {
					t.Fatal("LoanBrokerDelete detail is nil")
				}
				eqStr(t, "LoanBrokerId", v.LoanBrokerId, brokerID)
			},
		},
		{
			name: "LoanBrokerCoverDeposit",
			flat: map[string]any{
				"TransactionType": "LoanBrokerCoverDeposit",
				"Account":         testAccount,
				"LoanBrokerID":    brokerID,
				"Amount":          xrpAmount("1000000"),
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanBrokerCoverDeposit()
				if v == nil {
					t.Fatal("LoanBrokerCoverDeposit detail is nil")
				}
				eqStr(t, "LoanBrokerId", v.LoanBrokerId, brokerID)
				if v.Amount == nil {
					t.Fatal("Amount is nil")
				}
				eqStr(t, "Amount.Value", v.Amount.Value, "1000000")
			},
		},
		{
			name: "LoanBrokerCoverWithdraw",
			flat: map[string]any{
				"TransactionType": "LoanBrokerCoverWithdraw",
				"Account":         testAccount,
				"LoanBrokerID":    brokerID,
				"Amount":          xrpAmount("500"),
				"Destination":     testOther,
				"DestinationTag":  9,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanBrokerCoverWithdraw()
				if v == nil {
					t.Fatal("LoanBrokerCoverWithdraw detail is nil")
				}
				eqStr(t, "LoanBrokerId", v.LoanBrokerId, brokerID)
				if v.Amount == nil {
					t.Fatal("Amount is nil")
				}
				eqStr(t, "Amount.Value", v.Amount.Value, "500")
				eqStr(t, "Destination", v.Destination, testOther)
				eqU32(t, "DestinationTag", v.DestinationTag, 9)
			},
		},
		{
			name: "LoanBrokerCoverClawback",
			flat: map[string]any{
				"TransactionType": "LoanBrokerCoverClawback",
				"Account":         testAccount,
				"LoanBrokerID":    brokerID,
				"Amount":          xrpAmount("750"),
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanBrokerCoverClawback()
				if v == nil {
					t.Fatal("LoanBrokerCoverClawback detail is nil")
				}
				eqStr(t, "LoanBrokerId", v.LoanBrokerId, brokerID)
				if v.Amount == nil {
					t.Fatal("Amount is nil")
				}
				eqStr(t, "Amount.Value", v.Amount.Value, "750")
			},
		},
		{
			name: "LoanSet",
			flat: map[string]any{
				"TransactionType":         "LoanSet",
				"Account":                 testAccount,
				"LoanBrokerID":            brokerID,
				"Data":                    "ABCD",
				"Counterparty":            testIssuer,
				"LoanOriginationFee":      "100",
				"LoanServiceFee":          "200",
				"LatePaymentFee":          "300",
				"ClosePaymentFee":         "400",
				"OverpaymentFee":          10000,
				"InterestRate":            50000,
				"LateInterestRate":        60000,
				"CloseInterestRate":       70000,
				"OverpaymentInterestRate": 80000,
				"PrincipalRequested":      "1000000",
				"PaymentTotal":            12,
				"PaymentInterval":         2592000,
				"GracePeriod":             86400,
				"CounterpartySignature": map[string]any{
					"SigningPubKey": testPubKey,
					"TxnSignature":  testSig,
				},
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanSet()
				if v == nil {
					t.Fatal("LoanSet detail is nil")
				}
				eqStr(t, "LoanBrokerId", v.LoanBrokerId, brokerID)
				eqStr(t, "Data", v.Data, "ABCD")
				eqStr(t, "Counterparty", v.Counterparty, testIssuer)
				eqStr(t, "LoanOriginationFee", v.LoanOriginationFee, "100")
				eqStr(t, "LoanServiceFee", v.LoanServiceFee, "200")
				eqStr(t, "LatePaymentFee", v.LatePaymentFee, "300")
				eqStr(t, "ClosePaymentFee", v.ClosePaymentFee, "400")
				eqU32(t, "OverpaymentFee", v.OverpaymentFee, 10000)
				eqU32(t, "InterestRate", v.InterestRate, 50000)
				eqU32(t, "LateInterestRate", v.LateInterestRate, 60000)
				eqU32(t, "CloseInterestRate", v.CloseInterestRate, 70000)
				eqU32(t, "OverpaymentInterestRate", v.OverpaymentInterestRate, 80000)
				eqStr(t, "PrincipalRequested", v.PrincipalRequested, "1000000")
				eqU32(t, "PaymentTotal", v.PaymentTotal, 12)
				eqU32(t, "PaymentInterval", v.PaymentInterval, 2592000)
				eqU32(t, "GracePeriod", v.GracePeriod, 86400)
				if v.CounterpartySignature == nil {
					t.Fatal("CounterpartySignature is nil")
				}
				eqStr(t, "CounterpartySignature.SigningPubKey", v.CounterpartySignature.SigningPubKey, testPubKey)
				eqStr(t, "CounterpartySignature.TxnSignature", v.CounterpartySignature.TxnSignature, testSig)
			},
		},
		{
			name: "LoanSet_multisig_counterparty",
			flat: map[string]any{
				"TransactionType":    "LoanSet",
				"Account":            testAccount,
				"LoanBrokerID":       brokerID,
				"PrincipalRequested": "1000000",
				"CounterpartySignature": map[string]any{
					"Signers": []any{
						map[string]any{"Signer": map[string]any{
							"Account":       testIssuer,
							"SigningPubKey": testPubKey,
							"TxnSignature":  testSig,
						}},
					},
				},
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanSet()
				if v == nil {
					t.Fatal("LoanSet detail is nil")
				}
				if v.CounterpartySignature == nil {
					t.Fatal("CounterpartySignature is nil")
				}
				if len(v.CounterpartySignature.Signers) != 1 {
					t.Fatalf("Signers len = %d, want 1", len(v.CounterpartySignature.Signers))
				}
				s := v.CounterpartySignature.Signers[0]
				eqStr(t, "Signer.Account", s.Account, testIssuer)
				eqStr(t, "Signer.SigningPubKey", s.SigningPubKey, testPubKey)
				eqStr(t, "Signer.TxnSignature", s.TxnSignature, testSig)
			},
		},
		{
			name: "LoanDelete",
			flat: map[string]any{
				"TransactionType": "LoanDelete",
				"Account":         testAccount,
				"LoanID":          loanID,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanDelete()
				if v == nil {
					t.Fatal("LoanDelete detail is nil")
				}
				eqStr(t, "LoanId", v.LoanId, loanID)
			},
		},
		{
			name: "LoanManage",
			flat: map[string]any{
				"TransactionType": "LoanManage",
				"Account":         testAccount,
				"LoanID":          loanID,
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanManage()
				if v == nil {
					t.Fatal("LoanManage detail is nil")
				}
				eqStr(t, "LoanId", v.LoanId, loanID)
			},
		},
		{
			name: "LoanPay",
			flat: map[string]any{
				"TransactionType": "LoanPay",
				"Account":         testAccount,
				"LoanID":          loanID,
				"Amount":          xrpAmount("123456"),
			},
			verify: func(t *testing.T, tx *pbxrpl.Transaction) {
				v := tx.GetLoanPay()
				if v == nil {
					t.Fatal("LoanPay detail is nil")
				}
				eqStr(t, "LoanId", v.LoanId, loanID)
				if v.Amount == nil {
					t.Fatal("Amount is nil")
				}
				eqStr(t, "Amount.Value", v.Amount.Value, "123456")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.verify(t, mapToProto(t, tc.flat))
		})
	}
}
