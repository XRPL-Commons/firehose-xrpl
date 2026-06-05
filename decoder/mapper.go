package decoder

import (
	"fmt"
	"strconv"

	xrpltx "github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
	"go.uber.org/zap"
)

// Mapper handles mapping from goxrpl types to protobuf types
type Mapper struct {
	logger    *zap.Logger
	txMappers map[string]func(*pbxrpl.Transaction, xrpltx.FlatTransaction)
}

// NewMapper creates a new mapper with pre-built transaction type dispatch map
func NewMapper(logger *zap.Logger) *Mapper {
	m := &Mapper{
		logger: logger,
	}

	// Build hashmap for O(1) transaction type lookup instead of O(n) switch
	m.txMappers = map[string]func(*pbxrpl.Transaction, xrpltx.FlatTransaction){
		"Payment": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_Payment{Payment: m.mapPayment(flat)}
		},
		"OfferCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_OfferCreate{OfferCreate: m.mapOfferCreate(flat)}
		},
		"OfferCancel": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_OfferCancel{OfferCancel: m.mapOfferCancel(flat)}
		},
		"TrustSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_TrustSet{TrustSet: m.mapTrustSet(flat)}
		},
		"AccountSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AccountSet{AccountSet: m.mapAccountSet(flat)}
		},
		"AccountDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AccountDelete{AccountDelete: m.mapAccountDelete(flat)}
		},
		"SetRegularKey": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_SetRegularKey{SetRegularKey: m.mapSetRegularKey(flat)}
		},
		"SignerListSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_SignerListSet{SignerListSet: m.mapSignerListSet(flat)}
		},
		"EscrowCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_EscrowCreate{EscrowCreate: m.mapEscrowCreate(flat)}
		},
		"EscrowFinish": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_EscrowFinish{EscrowFinish: m.mapEscrowFinish(flat)}
		},
		"EscrowCancel": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_EscrowCancel{EscrowCancel: m.mapEscrowCancel(flat)}
		},
		"PaymentChannelCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_PaymentChannelCreate{PaymentChannelCreate: m.mapPaymentChannelCreate(flat)}
		},
		"PaymentChannelFund": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_PaymentChannelFund{PaymentChannelFund: m.mapPaymentChannelFund(flat)}
		},
		"PaymentChannelClaim": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_PaymentChannelClaim{PaymentChannelClaim: m.mapPaymentChannelClaim(flat)}
		},
		"CheckCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_CheckCreate{CheckCreate: m.mapCheckCreate(flat)}
		},
		"CheckCash": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_CheckCash{CheckCash: m.mapCheckCash(flat)}
		},
		"CheckCancel": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_CheckCancel{CheckCancel: m.mapCheckCancel(flat)}
		},
		"DepositPreauth": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_DepositPreauth{DepositPreauth: m.mapDepositPreauth(flat)}
		},
		"TicketCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_TicketCreate{TicketCreate: m.mapTicketCreate(flat)}
		},
		"NFTokenMint": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_NftokenMint{NftokenMint: m.mapNFTokenMint(flat)}
		},
		"NFTokenBurn": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_NftokenBurn{NftokenBurn: m.mapNFTokenBurn(flat)}
		},
		"NFTokenCreateOffer": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_NftokenCreateOffer{NftokenCreateOffer: m.mapNFTokenCreateOffer(flat)}
		},
		"NFTokenCancelOffer": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_NftokenCancelOffer{NftokenCancelOffer: m.mapNFTokenCancelOffer(flat)}
		},
		"NFTokenAcceptOffer": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_NftokenAcceptOffer{NftokenAcceptOffer: m.mapNFTokenAcceptOffer(flat)}
		},
		"Clawback": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_Clawback{Clawback: m.mapClawback(flat)}
		},
		"AMMCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AmmCreate{AmmCreate: m.mapAMMCreate(flat)}
		},
		"AMMDeposit": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AmmDeposit{AmmDeposit: m.mapAMMDeposit(flat)}
		},
		"AMMWithdraw": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AmmWithdraw{AmmWithdraw: m.mapAMMWithdraw(flat)}
		},
		"AMMVote": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AmmVote{AmmVote: m.mapAMMVote(flat)}
		},
		"AMMBid": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AmmBid{AmmBid: m.mapAMMBid(flat)}
		},
		"AMMDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AmmDelete{AmmDelete: m.mapAMMDelete(flat)}
		},
		"AMMClawback": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_AmmClawback{AmmClawback: m.mapAMMClawback(flat)}
		},
		"DIDSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_DidSet{DidSet: m.mapDIDSet(flat)}
		},
		"DIDDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_DidDelete{DidDelete: m.mapDIDDelete()}
		},
		"OracleSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_OracleSet{OracleSet: m.mapOracleSet(flat)}
		},
		"OracleDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_OracleDelete{OracleDelete: m.mapOracleDelete(flat)}
		},
		"MPTokenIssuanceCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_MptokenIssuanceCreate{MptokenIssuanceCreate: m.mapMPTokenIssuanceCreate(flat)}
		},
		"MPTokenIssuanceDestroy": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_MptokenIssuanceDestroy{MptokenIssuanceDestroy: m.mapMPTokenIssuanceDestroy(flat)}
		},
		"MPTokenIssuanceSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_MptokenIssuanceSet{MptokenIssuanceSet: m.mapMPTokenIssuanceSet(flat)}
		},
		"MPTokenAuthorize": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_MptokenAuthorize{MptokenAuthorize: m.mapMPTokenAuthorize(flat)}
		},
		"CredentialCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_CredentialCreate{CredentialCreate: m.mapCredentialCreate(flat)}
		},
		"CredentialAccept": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_CredentialAccept{CredentialAccept: m.mapCredentialAccept(flat)}
		},
		"CredentialDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_CredentialDelete{CredentialDelete: m.mapCredentialDelete(flat)}
		},
		"PermissionedDomainSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_PermissionedDomainSet{PermissionedDomainSet: m.mapPermissionedDomainSet(flat)}
		},
		"PermissionedDomainDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_PermissionedDomainDelete{PermissionedDomainDelete: m.mapPermissionedDomainDelete(flat)}
		},
		"DelegateSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_DelegateSet{DelegateSet: m.mapDelegateSet(flat)}
		},
		"Batch": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_Batch{Batch: m.mapBatch(flat)}
		},
		"VaultCreate": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_VaultCreate{VaultCreate: m.mapVaultCreate(flat)}
		},
		"VaultSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_VaultSet{VaultSet: m.mapVaultSet(flat)}
		},
		"VaultDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_VaultDelete{VaultDelete: m.mapVaultDelete(flat)}
		},
		"VaultDeposit": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_VaultDeposit{VaultDeposit: m.mapVaultDeposit(flat)}
		},
		"VaultWithdraw": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_VaultWithdraw{VaultWithdraw: m.mapVaultWithdraw(flat)}
		},
		"VaultClawback": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_VaultClawback{VaultClawback: m.mapVaultClawback(flat)}
		},
		"LoanBrokerSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanBrokerSet{LoanBrokerSet: m.mapLoanBrokerSet(flat)}
		},
		"LoanBrokerDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanBrokerDelete{LoanBrokerDelete: m.mapLoanBrokerDelete(flat)}
		},
		"LoanBrokerCoverDeposit": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanBrokerCoverDeposit{LoanBrokerCoverDeposit: m.mapLoanBrokerCoverDeposit(flat)}
		},
		"LoanBrokerCoverWithdraw": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanBrokerCoverWithdraw{LoanBrokerCoverWithdraw: m.mapLoanBrokerCoverWithdraw(flat)}
		},
		"LoanBrokerCoverClawback": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanBrokerCoverClawback{LoanBrokerCoverClawback: m.mapLoanBrokerCoverClawback(flat)}
		},
		"LoanSet": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanSet{LoanSet: m.mapLoanSet(flat)}
		},
		"LoanDelete": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanDelete{LoanDelete: m.mapLoanDelete(flat)}
		},
		"LoanManage": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanManage{LoanManage: m.mapLoanManage(flat)}
		},
		"LoanPay": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_LoanPay{LoanPay: m.mapLoanPay(flat)}
		},
		"EnableAmendment": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_EnableAmendment{EnableAmendment: m.mapEnableAmendment(flat)}
		},
		"SetFee": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_SetFee{SetFee: m.mapSetFee(flat)}
		},
		"UNLModify": func(tx *pbxrpl.Transaction, flat xrpltx.FlatTransaction) {
			tx.TxDetails = &pbxrpl.Transaction_UnlModify{UnlModify: m.mapUNLModify(flat)}
		},
		// XChain transactions removed - deprecated and will never be in prod
	}

	return m
}

// MapAmount converts goxrpl CurrencyAmount to protobuf Amount
func (m *Mapper) MapAmount(amt types.CurrencyAmount) *pbxrpl.Amount {
	if amt == nil {
		return nil
	}

	// Flatten the amount to get the underlying data
	flat := amt.Flatten()

	switch v := flat.(type) {
	case string:
		// XRP amount in drops
		return &pbxrpl.Amount{
			Value: v,
		}
	case map[string]interface{}:
		// Token or MPT amount
		result := &pbxrpl.Amount{}

		if value, ok := v["value"].(string); ok {
			result.Value = value
		}
		if currency, ok := v["currency"].(string); ok {
			result.Currency = currency
		}
		if issuer, ok := v["issuer"].(string); ok {
			result.Issuer = issuer
		}
		// Check for MPT issuance ID
		if mptID, ok := v["mpt_issuance_id"].(string); ok {
			result.MptIssuanceId = mptID
		}

		return result
	}

	return nil
}

// MapMemos converts goxrpl MemoWrapper array to protobuf Memo array
func (m *Mapper) MapMemos(memos []types.MemoWrapper) []*pbxrpl.Memo {
	if len(memos) == 0 {
		return nil
	}

	result := make([]*pbxrpl.Memo, 0, len(memos))
	for _, memoWrapper := range memos {
		flat := memoWrapper.Flatten()
		if memoMap, ok := flat["Memo"].(map[string]interface{}); ok {
			memo := &pbxrpl.Memo{}

			if data, ok := memoMap["MemoData"].(string); ok {
				memo.MemoData = data
			}
			if format, ok := memoMap["MemoFormat"].(string); ok {
				memo.MemoFormat = format
			}
			if memoType, ok := memoMap["MemoType"].(string); ok {
				memo.MemoType = memoType
			}

			result = append(result, memo)
		}
	}

	return result
}

// MapSigners converts goxrpl Signer array to protobuf Signer array
func (m *Mapper) MapSigners(signers []types.Signer) []*pbxrpl.Signer {
	if len(signers) == 0 {
		return nil
	}

	result := make([]*pbxrpl.Signer, 0, len(signers))
	for _, signer := range signers {
		flat := signer.Flatten()
		if signerMap, ok := flat["Signer"].(map[string]interface{}); ok {
			pbSigner := &pbxrpl.Signer{}

			if account, ok := signerMap["Account"].(string); ok {
				pbSigner.Account = account
			}
			if txnSig, ok := signerMap["TxnSignature"].(string); ok {
				pbSigner.TxnSignature = txnSig
			}
			if signingPubKey, ok := signerMap["SigningPubKey"].(string); ok {
				pbSigner.SigningPubKey = signingPubKey
			}

			result = append(result, pbSigner)
		}
	}

	return result
}

// MapBaseTxFields extracts common transaction fields from goxrpl BaseTx
func (m *Mapper) MapBaseTxFields(baseTx xrpltx.BaseTx) (account string, fee uint64, sequence uint32, flags uint32, memos []*pbxrpl.Memo, signers []*pbxrpl.Signer) {
	account = baseTx.Account.String()

	if baseTx.Fee != 0 {
		// Fee is in drops, convert XRPCurrencyAmount to uint64
		fee = uint64(baseTx.Fee)
	}

	sequence = baseTx.Sequence
	flags = baseTx.Flags

	memos = m.MapMemos(baseTx.Memos)
	signers = m.MapSigners(baseTx.Signers)

	return
}

// MapTransactionToProto maps a goxrpl FlatTransaction to protobuf Transaction
// This is the main entry point for mapping transaction data
func (m *Mapper) MapTransactionToProto(flatTx xrpltx.FlatTransaction, txBlob, metaBlob []byte, txHash []byte, txIndex uint32, result string) (*pbxrpl.Transaction, error) {
	// Extract transaction type
	txType, ok := flatTx["TransactionType"].(string)
	if !ok {
		return nil, fmt.Errorf("missing TransactionType in transaction")
	}

	// Extract common fields
	account := ""
	if acc, ok := flatTx["Account"].(string); ok {
		account = acc
	}

	fee := uint64(0)
	if feeStr, ok := flatTx["Fee"].(string); ok {
		// Fee is in drops as string, convert to uint64
		if feeVal, err := strconv.ParseUint(feeStr, 10, 64); err == nil {
			fee = feeVal
		}
	}

	sequence := uint32(0)
	if seq, ok := asUint32(flatTx["Sequence"]); ok {
		sequence = seq
	}

	flags := uint32(0)
	if f, ok := asUint32(flatTx["Flags"]); ok {
		flags = f
	}

	// Build base transaction
	protoTx := &pbxrpl.Transaction{
		Hash:     txHash,
		Result:   result,
		Index:    txIndex,
		TxBlob:   txBlob,
		MetaBlob: metaBlob,
		TxType:   txType,
		Account:  account,
		Fee:      fee,
		Sequence: sequence,
		Flags:    flags,
	}

	// Extract optional common fields
	if accountTxnID, ok := flatTx["AccountTxnID"].(string); ok {
		protoTx.AccountTxnId = accountTxnID
	}

	if delegate, ok := flatTx["Delegate"].(string); ok {
		protoTx.Delegate = delegate
	}

	if lastLedgerSeq, ok := asUint32(flatTx["LastLedgerSequence"]); ok {
		protoTx.LastLedgerSequence = lastLedgerSeq
	}

	// Map memos
	if memosRaw, ok := flatTx["Memos"].([]interface{}); ok {
		protoTx.Memos = m.mapMemosFromFlat(memosRaw)
	}

	if networkID, ok := asUint32(flatTx["NetworkID"]); ok {
		protoTx.NetworkId = networkID
	}

	// Map signers
	if signersRaw, ok := flatTx["Signers"].([]interface{}); ok {
		protoTx.Signers = m.mapSignersFromFlat(signersRaw)
	}

	if sourceTag, ok := asUint32(flatTx["SourceTag"]); ok {
		protoTx.SourceTag = sourceTag
	}

	if signingPubKey, ok := flatTx["SigningPubKey"].(string); ok {
		protoTx.SigningPubKey = signingPubKey
	}

	if ticketSeq, ok := asUint32(flatTx["TicketSequence"]); ok {
		protoTx.TicketSequence = ticketSeq
	}

	if txnSig, ok := flatTx["TxnSignature"].(string); ok {
		protoTx.TxnSignature = txnSig
	}

	// Map transaction-specific details based on type
	m.mapTxDetails(protoTx, flatTx, txType)

	return protoTx, nil
}

// Helper methods for mapping from flat representations

func (m *Mapper) mapMemosFromFlat(memosRaw []interface{}) []*pbxrpl.Memo {
	if len(memosRaw) == 0 {
		return nil
	}

	result := make([]*pbxrpl.Memo, 0, len(memosRaw))
	for _, memoRaw := range memosRaw {
		if memoMap, ok := memoRaw.(map[string]interface{}); ok {
			if memo, ok := memoMap["Memo"].(map[string]interface{}); ok {
				pbMemo := &pbxrpl.Memo{}

				if data, ok := memo["MemoData"].(string); ok {
					pbMemo.MemoData = data
				}
				if format, ok := memo["MemoFormat"].(string); ok {
					pbMemo.MemoFormat = format
				}
				if memoType, ok := memo["MemoType"].(string); ok {
					pbMemo.MemoType = memoType
				}

				result = append(result, pbMemo)
			}
		}
	}

	return result
}

func (m *Mapper) mapSignersFromFlat(signersRaw []interface{}) []*pbxrpl.Signer {
	if len(signersRaw) == 0 {
		return nil
	}

	result := make([]*pbxrpl.Signer, 0, len(signersRaw))
	for _, signerRaw := range signersRaw {
		if signerMap, ok := signerRaw.(map[string]interface{}); ok {
			if signer, ok := signerMap["Signer"].(map[string]interface{}); ok {
				pbSigner := &pbxrpl.Signer{}

				if account, ok := signer["Account"].(string); ok {
					pbSigner.Account = account
				}
				if txnSig, ok := signer["TxnSignature"].(string); ok {
					pbSigner.TxnSignature = txnSig
				}
				if signingPubKey, ok := signer["SigningPubKey"].(string); ok {
					pbSigner.SigningPubKey = signingPubKey
				}

				result = append(result, pbSigner)
			}
		}
	}

	return result
}

func (m *Mapper) mapAmountFromFlat(amtRaw interface{}) *pbxrpl.Amount {
	if amtRaw == nil {
		return nil
	}

	switch v := amtRaw.(type) {
	case string:
		// XRP amount in drops
		return &pbxrpl.Amount{
			Value: v,
		}
	case map[string]interface{}:
		// Token or MPT amount
		result := &pbxrpl.Amount{}

		if value, ok := v["value"].(string); ok {
			result.Value = value
		}
		if currency, ok := v["currency"].(string); ok {
			result.Currency = currency
		}
		if issuer, ok := v["issuer"].(string); ok {
			result.Issuer = issuer
		}
		if mptID, ok := v["mpt_issuance_id"].(string); ok {
			result.MptIssuanceId = mptID
		}

		return result
	}

	return nil
}

// mapTxDetails populates the tx_details oneof field based on transaction type
func (m *Mapper) mapTxDetails(tx *pbxrpl.Transaction, flatTx xrpltx.FlatTransaction, txType string) {
	// Use hashmap for O(1) lookup instead of O(n) switch with 50+ cases
	if mapper, ok := m.txMappers[txType]; ok {
		mapper(tx, flatTx)
	} else {
		// Unknown transaction type - log for debugging
		m.logger.Debug("unknown transaction type", zap.String("tx_type", txType))
	}
}

// Transaction-specific mappers

// Payment transactions
func (m *Mapper) mapPayment(flat xrpltx.FlatTransaction) *pbxrpl.Payment {
	payment := &pbxrpl.Payment{}

	if dest, ok := flat["Destination"].(string); ok {
		payment.Destination = dest
	}

	payment.Amount = m.mapAmountFromFlat(flat["Amount"])
	payment.DeliverMax = m.mapAmountFromFlat(flat["DeliverMax"])
	payment.SendMax = m.mapAmountFromFlat(flat["SendMax"])
	payment.DeliverMin = m.mapAmountFromFlat(flat["DeliverMin"])

	if invoiceID, ok := flat["InvoiceID"].(string); ok {
		payment.InvoiceId = invoiceID
	}

	if destTag, ok := asUint32(flat["DestinationTag"]); ok {
		payment.DestinationTag = destTag
	}

	if credIDs, ok := flat["CredentialIDs"].([]interface{}); ok {
		payment.CredentialIds = m.mapStringArray(credIDs)
	}

	if domainID, ok := flat["DomainID"].(string); ok {
		payment.DomainId = domainID
	}

	if paths, ok := flat["Paths"].([]interface{}); ok {
		payment.Paths = m.mapPaths(paths)
	}

	// DeliveredAmount from metadata would be mapped separately
	payment.DeliveredAmount = m.mapAmountFromFlat(flat["delivered_amount"])

	return payment
}

// DEX transactions
func (m *Mapper) mapOfferCreate(flat xrpltx.FlatTransaction) *pbxrpl.OfferCreate {
	offer := &pbxrpl.OfferCreate{}

	offer.TakerGets = m.mapAmountFromFlat(flat["TakerGets"])
	offer.TakerPays = m.mapAmountFromFlat(flat["TakerPays"])

	if exp, ok := asUint32(flat["Expiration"]); ok {
		offer.Expiration = exp
	}

	if offerSeq, ok := asUint32(flat["OfferSequence"]); ok {
		offer.OfferSequence = offerSeq
	}

	if domainID, ok := flat["DomainID"].(string); ok {
		offer.DomainId = domainID
	}

	return offer
}

func (m *Mapper) mapOfferCancel(flat xrpltx.FlatTransaction) *pbxrpl.OfferCancel {
	cancel := &pbxrpl.OfferCancel{}

	if offerSeq, ok := asUint32(flat["OfferSequence"]); ok {
		cancel.OfferSequence = offerSeq
	}

	return cancel
}

// Trustline
func (m *Mapper) mapTrustSet(flat xrpltx.FlatTransaction) *pbxrpl.TrustSet {
	trust := &pbxrpl.TrustSet{}

	trust.LimitAmount = m.mapAmountFromFlat(flat["LimitAmount"])

	if qualityIn, ok := asUint32(flat["QualityIn"]); ok {
		trust.QualityIn = qualityIn
	}

	if qualityOut, ok := asUint32(flat["QualityOut"]); ok {
		trust.QualityOut = qualityOut
	}

	return trust
}

// Account management
func (m *Mapper) mapAccountSet(flat xrpltx.FlatTransaction) *pbxrpl.AccountSet {
	acct := &pbxrpl.AccountSet{}

	if setFlag, ok := asUint32(flat["SetFlag"]); ok {
		acct.SetFlag = setFlag
	}

	if clearFlag, ok := asUint32(flat["ClearFlag"]); ok {
		acct.ClearFlag = clearFlag
	}

	if domain, ok := flat["Domain"].(string); ok {
		acct.Domain = domain
	}

	if emailHash, ok := flat["EmailHash"].(string); ok {
		acct.EmailHash = emailHash
	}

	if msgKey, ok := flat["MessageKey"].(string); ok {
		acct.MessageKey = msgKey
	}

	if transferRate, ok := asUint32(flat["TransferRate"]); ok {
		acct.TransferRate = transferRate
	}

	if tickSize, ok := asUint32(flat["TickSize"]); ok {
		acct.TickSize = tickSize
	}

	if minter, ok := flat["NFTokenMinter"].(string); ok {
		acct.NftokenMinter = minter
	}

	if walletLocator, ok := flat["WalletLocator"].(string); ok {
		acct.WalletLocator = walletLocator
	}

	if walletSize, ok := asUint32(flat["WalletSize"]); ok {
		acct.WalletSize = walletSize
	}

	return acct
}

func (m *Mapper) mapAccountDelete(flat xrpltx.FlatTransaction) *pbxrpl.AccountDelete {
	del := &pbxrpl.AccountDelete{}

	if dest, ok := flat["Destination"].(string); ok {
		del.Destination = dest
	}

	if destTag, ok := asUint32(flat["DestinationTag"]); ok {
		del.DestinationTag = destTag
	}

	if credIDs, ok := flat["CredentialIDs"].([]interface{}); ok {
		del.CredentialIds = m.mapStringArray(credIDs)
	}

	return del
}

func (m *Mapper) mapSetRegularKey(flat xrpltx.FlatTransaction) *pbxrpl.SetRegularKey {
	key := &pbxrpl.SetRegularKey{}

	if regKey, ok := flat["RegularKey"].(string); ok {
		key.RegularKey = regKey
	}

	return key
}

func (m *Mapper) mapSignerListSet(flat xrpltx.FlatTransaction) *pbxrpl.SignerListSet {
	sls := &pbxrpl.SignerListSet{}

	if quorum, ok := asUint32(flat["SignerQuorum"]); ok {
		sls.SignerQuorum = quorum
	}

	if entries, ok := flat["SignerEntries"].([]interface{}); ok {
		sls.SignerEntries = m.mapSignerEntries(entries)
	}

	return sls
}

// Escrow transactions
func (m *Mapper) mapEscrowCreate(flat xrpltx.FlatTransaction) *pbxrpl.EscrowCreate {
	escrow := &pbxrpl.EscrowCreate{}

	if dest, ok := flat["Destination"].(string); ok {
		escrow.Destination = dest
	}

	escrow.Amount = m.mapAmountFromFlat(flat["Amount"])

	if cancelAfter, ok := asUint32(flat["CancelAfter"]); ok {
		escrow.CancelAfter = cancelAfter
	}

	if finishAfter, ok := asUint32(flat["FinishAfter"]); ok {
		escrow.FinishAfter = finishAfter
	}

	if condition, ok := flat["Condition"].(string); ok {
		escrow.Condition = condition
	}

	if destTag, ok := asUint32(flat["DestinationTag"]); ok {
		escrow.DestinationTag = destTag
	}

	return escrow
}

func (m *Mapper) mapEscrowFinish(flat xrpltx.FlatTransaction) *pbxrpl.EscrowFinish {
	finish := &pbxrpl.EscrowFinish{}

	if owner, ok := flat["Owner"].(string); ok {
		finish.Owner = owner
	}

	if offerSeq, ok := asUint32(flat["OfferSequence"]); ok {
		finish.OfferSequence = offerSeq
	}

	if condition, ok := flat["Condition"].(string); ok {
		finish.Condition = condition
	}

	if fulfillment, ok := flat["Fulfillment"].(string); ok {
		finish.Fulfillment = fulfillment
	}

	if credIDs, ok := flat["CredentialIDs"].([]interface{}); ok {
		finish.CredentialIds = m.mapStringArray(credIDs)
	}

	return finish
}

func (m *Mapper) mapEscrowCancel(flat xrpltx.FlatTransaction) *pbxrpl.EscrowCancel {
	cancel := &pbxrpl.EscrowCancel{}

	if owner, ok := flat["Owner"].(string); ok {
		cancel.Owner = owner
	}

	if offerSeq, ok := asUint32(flat["OfferSequence"]); ok {
		cancel.OfferSequence = offerSeq
	}

	return cancel
}

// Payment channel transactions
func (m *Mapper) mapPaymentChannelCreate(flat xrpltx.FlatTransaction) *pbxrpl.PaymentChannelCreate {
	pc := &pbxrpl.PaymentChannelCreate{}

	if dest, ok := flat["Destination"].(string); ok {
		pc.Destination = dest
	}

	pc.Amount = m.mapAmountFromFlat(flat["Amount"])

	if settleDelay, ok := asUint32(flat["SettleDelay"]); ok {
		pc.SettleDelay = settleDelay
	}

	if pubKey, ok := flat["PublicKey"].(string); ok {
		pc.PublicKey = pubKey
	}

	if cancelAfter, ok := asUint32(flat["CancelAfter"]); ok {
		pc.CancelAfter = cancelAfter
	}

	if destTag, ok := asUint32(flat["DestinationTag"]); ok {
		pc.DestinationTag = destTag
	}

	return pc
}

func (m *Mapper) mapPaymentChannelFund(flat xrpltx.FlatTransaction) *pbxrpl.PaymentChannelFund {
	fund := &pbxrpl.PaymentChannelFund{}

	if channel, ok := flat["Channel"].(string); ok {
		fund.Channel = channel
	}

	fund.Amount = m.mapAmountFromFlat(flat["Amount"])

	if expiration, ok := asUint32(flat["Expiration"]); ok {
		fund.Expiration = expiration
	}

	return fund
}

func (m *Mapper) mapPaymentChannelClaim(flat xrpltx.FlatTransaction) *pbxrpl.PaymentChannelClaim {
	claim := &pbxrpl.PaymentChannelClaim{}

	if channel, ok := flat["Channel"].(string); ok {
		claim.Channel = channel
	}

	claim.Amount = m.mapAmountFromFlat(flat["Amount"])
	claim.Balance = m.mapAmountFromFlat(flat["Balance"])

	if signature, ok := flat["Signature"].(string); ok {
		claim.Signature = signature
	}

	if pubKey, ok := flat["PublicKey"].(string); ok {
		claim.PublicKey = pubKey
	}

	if credIDs, ok := flat["CredentialIDs"].([]interface{}); ok {
		claim.CredentialIds = m.mapStringArray(credIDs)
	}

	return claim
}

// Check transactions
func (m *Mapper) mapCheckCreate(flat xrpltx.FlatTransaction) *pbxrpl.CheckCreate {
	check := &pbxrpl.CheckCreate{}

	if dest, ok := flat["Destination"].(string); ok {
		check.Destination = dest
	}

	check.SendMax = m.mapAmountFromFlat(flat["SendMax"])

	if expiration, ok := asUint32(flat["Expiration"]); ok {
		check.Expiration = expiration
	}

	if destTag, ok := asUint32(flat["DestinationTag"]); ok {
		check.DestinationTag = destTag
	}

	if invoiceID, ok := flat["InvoiceID"].(string); ok {
		check.InvoiceId = invoiceID
	}

	return check
}

func (m *Mapper) mapCheckCash(flat xrpltx.FlatTransaction) *pbxrpl.CheckCash {
	cash := &pbxrpl.CheckCash{}

	if checkID, ok := flat["CheckID"].(string); ok {
		cash.CheckId = checkID
	}

	cash.Amount = m.mapAmountFromFlat(flat["Amount"])
	cash.DeliverMin = m.mapAmountFromFlat(flat["DeliverMin"])

	return cash
}

func (m *Mapper) mapCheckCancel(flat xrpltx.FlatTransaction) *pbxrpl.CheckCancel {
	cancel := &pbxrpl.CheckCancel{}

	if checkID, ok := flat["CheckID"].(string); ok {
		cancel.CheckId = checkID
	}

	return cancel
}

// Deposit preauth
func (m *Mapper) mapDepositPreauth(flat xrpltx.FlatTransaction) *pbxrpl.DepositPreauth {
	dp := &pbxrpl.DepositPreauth{}

	if auth, ok := flat["Authorize"].(string); ok {
		dp.Authorize = auth
	}

	if unauth, ok := flat["Unauthorize"].(string); ok {
		dp.Unauthorize = unauth
	}

	// TODO: Map AuthorizeCredentials and UnauthorizeCredentials arrays

	return dp
}

// Ticket
func (m *Mapper) mapTicketCreate(flat xrpltx.FlatTransaction) *pbxrpl.TicketCreate {
	ticket := &pbxrpl.TicketCreate{}

	if count, ok := asUint32(flat["TicketCount"]); ok {
		ticket.TicketCount = count
	}

	return ticket
}

// NFT transactions
func (m *Mapper) mapNFTokenMint(flat xrpltx.FlatTransaction) *pbxrpl.NFTokenMint {
	mint := &pbxrpl.NFTokenMint{}

	if taxon, ok := asUint32(flat["NFTokenTaxon"]); ok {
		mint.NftokenTaxon = taxon
	}

	if issuer, ok := flat["Issuer"].(string); ok {
		mint.Issuer = issuer
	}

	if transferFee, ok := asUint32(flat["TransferFee"]); ok {
		mint.TransferFee = transferFee
	}

	if uri, ok := flat["URI"].(string); ok {
		mint.Uri = uri
	}

	mint.Amount = m.mapAmountFromFlat(flat["Amount"])

	if expiration, ok := asUint32(flat["Expiration"]); ok {
		mint.Expiration = expiration
	}

	if dest, ok := flat["Destination"].(string); ok {
		mint.Destination = dest
	}

	return mint
}

func (m *Mapper) mapNFTokenBurn(flat xrpltx.FlatTransaction) *pbxrpl.NFTokenBurn {
	burn := &pbxrpl.NFTokenBurn{}

	if tokenID, ok := flat["NFTokenID"].(string); ok {
		burn.NftokenId = tokenID
	}

	if owner, ok := flat["Owner"].(string); ok {
		burn.Owner = owner
	}

	return burn
}

func (m *Mapper) mapNFTokenCreateOffer(flat xrpltx.FlatTransaction) *pbxrpl.NFTokenCreateOffer {
	offer := &pbxrpl.NFTokenCreateOffer{}

	if tokenID, ok := flat["NFTokenID"].(string); ok {
		offer.NftokenId = tokenID
	}

	offer.Amount = m.mapAmountFromFlat(flat["Amount"])

	if owner, ok := flat["Owner"].(string); ok {
		offer.Owner = owner
	}

	if dest, ok := flat["Destination"].(string); ok {
		offer.Destination = dest
	}

	if expiration, ok := asUint32(flat["Expiration"]); ok {
		offer.Expiration = expiration
	}

	return offer
}

func (m *Mapper) mapNFTokenCancelOffer(flat xrpltx.FlatTransaction) *pbxrpl.NFTokenCancelOffer {
	cancel := &pbxrpl.NFTokenCancelOffer{}

	if offers, ok := flat["NFTokenOffers"].([]interface{}); ok {
		cancel.NftokenOffers = m.mapStringArray(offers)
	}

	return cancel
}

func (m *Mapper) mapNFTokenAcceptOffer(flat xrpltx.FlatTransaction) *pbxrpl.NFTokenAcceptOffer {
	accept := &pbxrpl.NFTokenAcceptOffer{}

	if sellOffer, ok := flat["NFTokenSellOffer"].(string); ok {
		accept.NftokenSellOffer = sellOffer
	}

	if buyOffer, ok := flat["NFTokenBuyOffer"].(string); ok {
		accept.NftokenBuyOffer = buyOffer
	}

	accept.NftokenBrokerFee = m.mapAmountFromFlat(flat["NFTokenBrokerFee"])

	return accept
}

// Clawback
func (m *Mapper) mapClawback(flat xrpltx.FlatTransaction) *pbxrpl.Clawback {
	clawback := &pbxrpl.Clawback{}

	clawback.Amount = m.mapAmountFromFlat(flat["Amount"])

	if holder, ok := flat["Holder"].(string); ok {
		clawback.Holder = holder
	}

	return clawback
}

// AMM transactions
func (m *Mapper) mapAMMCreate(flat xrpltx.FlatTransaction) *pbxrpl.AMMCreate {
	amm := &pbxrpl.AMMCreate{}

	amm.Amount = m.mapAmountFromFlat(flat["Amount"])
	amm.Amount2 = m.mapAmountFromFlat(flat["Amount2"])

	if tradingFee, ok := asUint32(flat["TradingFee"]); ok {
		amm.TradingFee = tradingFee
	}

	return amm
}

func (m *Mapper) mapAMMDeposit(flat xrpltx.FlatTransaction) *pbxrpl.AMMDeposit {
	deposit := &pbxrpl.AMMDeposit{}

	deposit.Asset = m.mapAssetFromFlat(flat["Asset"])
	deposit.Asset2 = m.mapAssetFromFlat(flat["Asset2"])
	deposit.Amount = m.mapAmountFromFlat(flat["Amount"])
	deposit.Amount2 = m.mapAmountFromFlat(flat["Amount2"])
	deposit.EPrice = m.mapAmountFromFlat(flat["EPrice"])
	deposit.LpTokenOut = m.mapAmountFromFlat(flat["LPTokenOut"])

	if tradingFee, ok := asUint32(flat["TradingFee"]); ok {
		deposit.TradingFee = tradingFee
	}

	return deposit
}

func (m *Mapper) mapAMMWithdraw(flat xrpltx.FlatTransaction) *pbxrpl.AMMWithdraw {
	withdraw := &pbxrpl.AMMWithdraw{}

	withdraw.Asset = m.mapAssetFromFlat(flat["Asset"])
	withdraw.Asset2 = m.mapAssetFromFlat(flat["Asset2"])
	withdraw.Amount = m.mapAmountFromFlat(flat["Amount"])
	withdraw.Amount2 = m.mapAmountFromFlat(flat["Amount2"])
	withdraw.EPrice = m.mapAmountFromFlat(flat["EPrice"])
	withdraw.LpTokenIn = m.mapAmountFromFlat(flat["LPTokenIn"])

	return withdraw
}

func (m *Mapper) mapAMMVote(flat xrpltx.FlatTransaction) *pbxrpl.AMMVote {
	vote := &pbxrpl.AMMVote{}

	vote.Asset = m.mapAssetFromFlat(flat["Asset"])
	vote.Asset2 = m.mapAssetFromFlat(flat["Asset2"])

	if tradingFee, ok := asUint32(flat["TradingFee"]); ok {
		vote.TradingFee = tradingFee
	}

	return vote
}

func (m *Mapper) mapAMMBid(flat xrpltx.FlatTransaction) *pbxrpl.AMMBid {
	bid := &pbxrpl.AMMBid{}

	bid.Asset = m.mapAssetFromFlat(flat["Asset"])
	bid.Asset2 = m.mapAssetFromFlat(flat["Asset2"])
	bid.BidMin = m.mapAmountFromFlat(flat["BidMin"])
	bid.BidMax = m.mapAmountFromFlat(flat["BidMax"])

	if authAccounts, ok := flat["AuthAccounts"].([]interface{}); ok {
		bid.AuthAccounts = m.mapAuthAccounts(authAccounts)
	}

	return bid
}

func (m *Mapper) mapAMMDelete(flat xrpltx.FlatTransaction) *pbxrpl.AMMDelete {
	del := &pbxrpl.AMMDelete{}

	del.Asset = m.mapAssetFromFlat(flat["Asset"])
	del.Asset2 = m.mapAssetFromFlat(flat["Asset2"])

	return del
}

func (m *Mapper) mapAMMClawback(flat xrpltx.FlatTransaction) *pbxrpl.AMMClawback {
	clawback := &pbxrpl.AMMClawback{}

	if holder, ok := flat["Holder"].(string); ok {
		clawback.Holder = holder
	}

	clawback.Asset = m.mapAssetFromFlat(flat["Asset"])
	clawback.Asset2 = m.mapAssetFromFlat(flat["Asset2"])
	clawback.Amount = m.mapAmountFromFlat(flat["Amount"])

	return clawback
}

// DID transactions
func (m *Mapper) mapDIDSet(flat xrpltx.FlatTransaction) *pbxrpl.DIDSet {
	did := &pbxrpl.DIDSet{}

	if didDoc, ok := flat["DIDDocument"].(string); ok {
		did.DidDocument = didDoc
	}

	if uri, ok := flat["URI"].(string); ok {
		did.Uri = uri
	}

	if data, ok := flat["Data"].(string); ok {
		did.Data = data
	}

	return did
}

func (m *Mapper) mapDIDDelete() *pbxrpl.DIDDelete {
	return &pbxrpl.DIDDelete{}
}

// Oracle transactions
func (m *Mapper) mapOracleSet(flat xrpltx.FlatTransaction) *pbxrpl.OracleSet {
	oracle := &pbxrpl.OracleSet{}

	if oracleDocID, ok := asUint32(flat["OracleDocumentID"]); ok {
		oracle.OracleDocumentId = oracleDocID
	}

	if provider, ok := flat["Provider"].(string); ok {
		oracle.Provider = provider
	}

	if assetClass, ok := flat["AssetClass"].(string); ok {
		oracle.AssetClass = assetClass
	}

	if lastUpdateTime, ok := asUint32(flat["LastUpdateTime"]); ok {
		oracle.LastUpdateTime = lastUpdateTime
	}

	if priceDataSeries, ok := flat["PriceDataSeries"].([]interface{}); ok {
		oracle.PriceDataSeries = m.mapPriceDataSeries(priceDataSeries)
	}

	return oracle
}

func (m *Mapper) mapOracleDelete(flat xrpltx.FlatTransaction) *pbxrpl.OracleDelete {
	del := &pbxrpl.OracleDelete{}

	if oracleDocID, ok := asUint32(flat["OracleDocumentID"]); ok {
		del.OracleDocumentId = oracleDocID
	}

	return del
}

// MPToken transactions
func (m *Mapper) mapMPTokenIssuanceCreate(flat xrpltx.FlatTransaction) *pbxrpl.MPTokenIssuanceCreate {
	create := &pbxrpl.MPTokenIssuanceCreate{}

	if assetScale, ok := asUint32(flat["AssetScale"]); ok {
		create.AssetScale = assetScale
	}

	if maxAmount, ok := flat["MaximumAmount"].(string); ok {
		if amount, err := strconv.ParseUint(maxAmount, 10, 64); err == nil {
			create.MaximumAmount = amount
		}
	}

	if transferFee, ok := asUint32(flat["TransferFee"]); ok {
		create.TransferFee = transferFee
	}

	if metadata, ok := flat["MPTokenMetadata"].(string); ok {
		create.MptokenMetadata = metadata
	}

	return create
}

func (m *Mapper) mapMPTokenIssuanceDestroy(flat xrpltx.FlatTransaction) *pbxrpl.MPTokenIssuanceDestroy {
	destroy := &pbxrpl.MPTokenIssuanceDestroy{}

	if issuanceID, ok := flat["MPTokenIssuanceID"].(string); ok {
		destroy.MptokenIssuanceId = issuanceID
	}

	return destroy
}

func (m *Mapper) mapMPTokenIssuanceSet(flat xrpltx.FlatTransaction) *pbxrpl.MPTokenIssuanceSet {
	set := &pbxrpl.MPTokenIssuanceSet{}

	if issuanceID, ok := flat["MPTokenIssuanceID"].(string); ok {
		set.MptokenIssuanceId = issuanceID
	}

	if holder, ok := flat["Holder"].(string); ok {
		set.Holder = holder
	}

	return set
}

func (m *Mapper) mapMPTokenAuthorize(flat xrpltx.FlatTransaction) *pbxrpl.MPTokenAuthorize {
	auth := &pbxrpl.MPTokenAuthorize{}

	if issuanceID, ok := flat["MPTokenIssuanceID"].(string); ok {
		auth.MptokenIssuanceId = issuanceID
	}

	if holder, ok := flat["Holder"].(string); ok {
		auth.Holder = holder
	}

	return auth
}

// Credential transactions
func (m *Mapper) mapCredentialCreate(flat xrpltx.FlatTransaction) *pbxrpl.CredentialCreate {
	cred := &pbxrpl.CredentialCreate{}

	if subject, ok := flat["Subject"].(string); ok {
		cred.Subject = subject
	}

	if credType, ok := flat["CredentialType"].(string); ok {
		cred.CredentialType = credType
	}

	if uri, ok := flat["URI"].(string); ok {
		cred.Uri = uri
	}

	if expiration, ok := asUint32(flat["Expiration"]); ok {
		cred.Expiration = expiration
	}

	return cred
}

func (m *Mapper) mapCredentialAccept(flat xrpltx.FlatTransaction) *pbxrpl.CredentialAccept {
	accept := &pbxrpl.CredentialAccept{}

	if issuer, ok := flat["Issuer"].(string); ok {
		accept.Issuer = issuer
	}

	if credType, ok := flat["CredentialType"].(string); ok {
		accept.CredentialType = credType
	}

	return accept
}

func (m *Mapper) mapCredentialDelete(flat xrpltx.FlatTransaction) *pbxrpl.CredentialDelete {
	del := &pbxrpl.CredentialDelete{}

	if subject, ok := flat["Subject"].(string); ok {
		del.Subject = subject
	}

	if credType, ok := flat["CredentialType"].(string); ok {
		del.CredentialType = credType
	}

	return del
}

// Permissioned domain transactions
func (m *Mapper) mapPermissionedDomainSet(flat xrpltx.FlatTransaction) *pbxrpl.PermissionedDomainSet {
	set := &pbxrpl.PermissionedDomainSet{}

	if domain, ok := flat["Domain"].(string); ok {
		set.DomainId = domain
	}

	return set
}

func (m *Mapper) mapPermissionedDomainDelete(flat xrpltx.FlatTransaction) *pbxrpl.PermissionedDomainDelete {
	del := &pbxrpl.PermissionedDomainDelete{}

	if domain, ok := flat["Domain"].(string); ok {
		del.DomainId = domain
	}

	return del
}

// Delegate transactions
func (m *Mapper) mapDelegateSet(flat xrpltx.FlatTransaction) *pbxrpl.DelegateSet {
	delegate := &pbxrpl.DelegateSet{}

	if del, ok := flat["Delegate"].(string); ok {
		delegate.Authorize = del
	}

	return delegate
}

// Batch transaction
func (m *Mapper) mapBatch(flat xrpltx.FlatTransaction) *pbxrpl.Batch {
	batch := &pbxrpl.Batch{}

	if rawTxs, ok := flat["RawTransactions"].([]interface{}); ok {
		batch.RawTransactions = m.mapRawTransactions(rawTxs)
	}

	if batchSigners, ok := flat["BatchSigners"].([]interface{}); ok {
		batch.BatchSigners = m.mapBatchSigners(batchSigners)
	}

	return batch
}

// System transactions (pseudo-transactions)
func (m *Mapper) mapEnableAmendment(flat xrpltx.FlatTransaction) *pbxrpl.EnableAmendment {
	amend := &pbxrpl.EnableAmendment{}

	if amendment, ok := flat["Amendment"].(string); ok {
		amend.Amendment = amendment
	}

	if ledgerSeq, ok := asUint32(flat["LedgerSequence"]); ok {
		amend.LedgerSequence = ledgerSeq
	}

	return amend
}

func (m *Mapper) mapSetFee(flat xrpltx.FlatTransaction) *pbxrpl.SetFee {
	fee := &pbxrpl.SetFee{}

	if baseFee, ok := flat["BaseFee"].(string); ok {
		if parsed, err := strconv.ParseUint(baseFee, 10, 64); err == nil {
			fee.BaseFee = parsed
		}
	}

	if refFeeUnits, ok := asUint32(flat["ReferenceFeeUnits"]); ok {
		fee.ReferenceFeeUnits = refFeeUnits
	}

	if reserveBase, ok := asUint32(flat["ReserveBase"]); ok {
		fee.ReserveBase = reserveBase
	}

	if reserveInc, ok := asUint32(flat["ReserveIncrement"]); ok {
		fee.ReserveIncrement = reserveInc
	}

	if ledgerSeq, ok := asUint32(flat["LedgerSequence"]); ok {
		fee.LedgerSequence = ledgerSeq
	}

	return fee
}

func (m *Mapper) mapUNLModify(flat xrpltx.FlatTransaction) *pbxrpl.UNLModify {
	unl := &pbxrpl.UNLModify{}

	if ledgerSeq, ok := asUint32(flat["LedgerSequence"]); ok {
		unl.LedgerSequence = ledgerSeq
	}

	if unlModifyDisabling, ok := asUint32(flat["UNLModifyDisabling"]); ok {
		unl.UnlModifyDisabling = unlModifyDisabling != 0
	}

	if unlModifyValidator, ok := flat["UNLModifyValidator"].(string); ok {
		unl.UnlModifyValidator = unlModifyValidator
	}

	return unl
}

// Vault transactions (single-asset vault)

func (m *Mapper) mapVaultCreate(flat xrpltx.FlatTransaction) *pbxrpl.VaultCreate {
	create := &pbxrpl.VaultCreate{}

	create.Asset = m.mapAssetFromFlat(flat["Asset"])

	if assetsMax, ok := flat["AssetsMaximum"].(string); ok {
		create.AssetsMaximum = assetsMax
	}

	if metadata, ok := flat["MPTokenMetadata"].(string); ok {
		create.MptokenMetadata = metadata
	}

	if domainID, ok := flat["DomainID"].(string); ok {
		create.DomainId = domainID
	}

	if policy, ok := asUint32(flat["WithdrawalPolicy"]); ok {
		create.WithdrawalPolicy = policy
	}

	if data, ok := flat["Data"].(string); ok {
		create.Data = data
	}

	if scale, ok := asUint32(flat["Scale"]); ok {
		create.Scale = scale
	}

	return create
}

func (m *Mapper) mapVaultSet(flat xrpltx.FlatTransaction) *pbxrpl.VaultSet {
	set := &pbxrpl.VaultSet{}

	if vaultID, ok := flat["VaultID"].(string); ok {
		set.VaultId = vaultID
	}

	if assetsMax, ok := flat["AssetsMaximum"].(string); ok {
		set.AssetsMaximum = assetsMax
	}

	if domainID, ok := flat["DomainID"].(string); ok {
		set.DomainId = domainID
	}

	if data, ok := flat["Data"].(string); ok {
		set.Data = data
	}

	return set
}

func (m *Mapper) mapVaultDelete(flat xrpltx.FlatTransaction) *pbxrpl.VaultDelete {
	del := &pbxrpl.VaultDelete{}

	if vaultID, ok := flat["VaultID"].(string); ok {
		del.VaultId = vaultID
	}

	return del
}

func (m *Mapper) mapVaultDeposit(flat xrpltx.FlatTransaction) *pbxrpl.VaultDeposit {
	deposit := &pbxrpl.VaultDeposit{}

	if vaultID, ok := flat["VaultID"].(string); ok {
		deposit.VaultId = vaultID
	}

	deposit.Amount = m.mapAmountFromFlat(flat["Amount"])

	return deposit
}

func (m *Mapper) mapVaultWithdraw(flat xrpltx.FlatTransaction) *pbxrpl.VaultWithdraw {
	withdraw := &pbxrpl.VaultWithdraw{}

	if vaultID, ok := flat["VaultID"].(string); ok {
		withdraw.VaultId = vaultID
	}

	withdraw.Amount = m.mapAmountFromFlat(flat["Amount"])

	if dest, ok := flat["Destination"].(string); ok {
		withdraw.Destination = dest
	}

	if destTag, ok := asUint32(flat["DestinationTag"]); ok {
		withdraw.DestinationTag = destTag
	}

	return withdraw
}

func (m *Mapper) mapVaultClawback(flat xrpltx.FlatTransaction) *pbxrpl.VaultClawback {
	clawback := &pbxrpl.VaultClawback{}

	if vaultID, ok := flat["VaultID"].(string); ok {
		clawback.VaultId = vaultID
	}

	if holder, ok := flat["Holder"].(string); ok {
		clawback.Holder = holder
	}

	clawback.Amount = m.mapAmountFromFlat(flat["Amount"])

	return clawback
}

// Lending transactions (lending protocol)

func (m *Mapper) mapLoanBrokerSet(flat xrpltx.FlatTransaction) *pbxrpl.LoanBrokerSet {
	set := &pbxrpl.LoanBrokerSet{}

	if vaultID, ok := flat["VaultID"].(string); ok {
		set.VaultId = vaultID
	}

	if brokerID, ok := flat["LoanBrokerID"].(string); ok {
		set.LoanBrokerId = brokerID
	}

	if data, ok := flat["Data"].(string); ok {
		set.Data = data
	}

	if feeRate, ok := asUint32(flat["ManagementFeeRate"]); ok {
		set.ManagementFeeRate = feeRate
	}

	if debtMax, ok := flat["DebtMaximum"].(string); ok {
		set.DebtMaximum = debtMax
	}

	if coverMin, ok := asUint32(flat["CoverRateMinimum"]); ok {
		set.CoverRateMinimum = coverMin
	}

	if coverLiq, ok := asUint32(flat["CoverRateLiquidation"]); ok {
		set.CoverRateLiquidation = coverLiq
	}

	return set
}

func (m *Mapper) mapLoanBrokerDelete(flat xrpltx.FlatTransaction) *pbxrpl.LoanBrokerDelete {
	del := &pbxrpl.LoanBrokerDelete{}

	if brokerID, ok := flat["LoanBrokerID"].(string); ok {
		del.LoanBrokerId = brokerID
	}

	return del
}

func (m *Mapper) mapLoanBrokerCoverDeposit(flat xrpltx.FlatTransaction) *pbxrpl.LoanBrokerCoverDeposit {
	deposit := &pbxrpl.LoanBrokerCoverDeposit{}

	if brokerID, ok := flat["LoanBrokerID"].(string); ok {
		deposit.LoanBrokerId = brokerID
	}

	deposit.Amount = m.mapAmountFromFlat(flat["Amount"])

	return deposit
}

func (m *Mapper) mapLoanBrokerCoverWithdraw(flat xrpltx.FlatTransaction) *pbxrpl.LoanBrokerCoverWithdraw {
	withdraw := &pbxrpl.LoanBrokerCoverWithdraw{}

	if brokerID, ok := flat["LoanBrokerID"].(string); ok {
		withdraw.LoanBrokerId = brokerID
	}

	withdraw.Amount = m.mapAmountFromFlat(flat["Amount"])

	if dest, ok := flat["Destination"].(string); ok {
		withdraw.Destination = dest
	}

	if destTag, ok := asUint32(flat["DestinationTag"]); ok {
		withdraw.DestinationTag = destTag
	}

	return withdraw
}

func (m *Mapper) mapLoanBrokerCoverClawback(flat xrpltx.FlatTransaction) *pbxrpl.LoanBrokerCoverClawback {
	clawback := &pbxrpl.LoanBrokerCoverClawback{}

	if brokerID, ok := flat["LoanBrokerID"].(string); ok {
		clawback.LoanBrokerId = brokerID
	}

	clawback.Amount = m.mapAmountFromFlat(flat["Amount"])

	return clawback
}

func (m *Mapper) mapLoanSet(flat xrpltx.FlatTransaction) *pbxrpl.LoanSet {
	loan := &pbxrpl.LoanSet{}

	if brokerID, ok := flat["LoanBrokerID"].(string); ok {
		loan.LoanBrokerId = brokerID
	}

	if data, ok := flat["Data"].(string); ok {
		loan.Data = data
	}

	if counterparty, ok := flat["Counterparty"].(string); ok {
		loan.Counterparty = counterparty
	}

	if sig, ok := flat["CounterpartySignature"].(map[string]interface{}); ok {
		loan.CounterpartySignature = m.mapCounterpartySignature(sig)
	}

	if fee, ok := flat["LoanOriginationFee"].(string); ok {
		loan.LoanOriginationFee = fee
	}

	if fee, ok := flat["LoanServiceFee"].(string); ok {
		loan.LoanServiceFee = fee
	}

	if fee, ok := flat["LatePaymentFee"].(string); ok {
		loan.LatePaymentFee = fee
	}

	if fee, ok := flat["ClosePaymentFee"].(string); ok {
		loan.ClosePaymentFee = fee
	}

	if rate, ok := asUint32(flat["OverpaymentFee"]); ok {
		loan.OverpaymentFee = rate
	}

	if rate, ok := asUint32(flat["InterestRate"]); ok {
		loan.InterestRate = rate
	}

	if rate, ok := asUint32(flat["LateInterestRate"]); ok {
		loan.LateInterestRate = rate
	}

	if rate, ok := asUint32(flat["CloseInterestRate"]); ok {
		loan.CloseInterestRate = rate
	}

	if rate, ok := asUint32(flat["OverpaymentInterestRate"]); ok {
		loan.OverpaymentInterestRate = rate
	}

	if principal, ok := flat["PrincipalRequested"].(string); ok {
		loan.PrincipalRequested = principal
	}

	if total, ok := asUint32(flat["PaymentTotal"]); ok {
		loan.PaymentTotal = total
	}

	if interval, ok := asUint32(flat["PaymentInterval"]); ok {
		loan.PaymentInterval = interval
	}

	if grace, ok := asUint32(flat["GracePeriod"]); ok {
		loan.GracePeriod = grace
	}

	return loan
}

func (m *Mapper) mapLoanDelete(flat xrpltx.FlatTransaction) *pbxrpl.LoanDelete {
	del := &pbxrpl.LoanDelete{}

	if loanID, ok := flat["LoanID"].(string); ok {
		del.LoanId = loanID
	}

	return del
}

func (m *Mapper) mapLoanManage(flat xrpltx.FlatTransaction) *pbxrpl.LoanManage {
	manage := &pbxrpl.LoanManage{}

	if loanID, ok := flat["LoanID"].(string); ok {
		manage.LoanId = loanID
	}

	return manage
}

func (m *Mapper) mapLoanPay(flat xrpltx.FlatTransaction) *pbxrpl.LoanPay {
	pay := &pbxrpl.LoanPay{}

	if loanID, ok := flat["LoanID"].(string); ok {
		pay.LoanId = loanID
	}

	pay.Amount = m.mapAmountFromFlat(flat["Amount"])

	return pay
}

// mapCounterpartySignature maps the inner counterparty signature object of a
// LoanSet (single-signature or multi-signature form).
func (m *Mapper) mapCounterpartySignature(sig map[string]interface{}) *pbxrpl.CounterpartySignature {
	cs := &pbxrpl.CounterpartySignature{}

	if pubKey, ok := sig["SigningPubKey"].(string); ok {
		cs.SigningPubKey = pubKey
	}

	if txnSig, ok := sig["TxnSignature"].(string); ok {
		cs.TxnSignature = txnSig
	}

	if signers, ok := sig["Signers"].([]interface{}); ok {
		cs.Signers = m.mapSignersFromFlat(signers)
	}

	return cs
}

// Helper functions for complex nested structures

// asUint32 extracts an unsigned 32-bit integer from a decoded transaction
// field. The XRPL binary codec (binarycodec.Decode) decodes UInt8 and UInt16
// fields as int and UInt32 fields as uint32 — never float64 — so a plain
// float64 type assertion silently drops every integer field. float64 is still
// accepted here for robustness against any JSON-sourced values.
func asUint32(v interface{}) (uint32, bool) {
	switch n := v.(type) {
	case uint32:
		return n, true
	case int:
		return uint32(n), true
	case int32:
		return uint32(n), true
	case int64:
		return uint32(n), true
	case uint64:
		return uint32(n), true
	case float64:
		return uint32(n), true
	default:
		return 0, false
	}
}

func (m *Mapper) mapStringArray(arr []interface{}) []string {
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

func (m *Mapper) mapPaths(pathsRaw []interface{}) []*pbxrpl.Path {
	result := make([]*pbxrpl.Path, 0, len(pathsRaw))
	for _, pathRaw := range pathsRaw {
		if pathArr, ok := pathRaw.([]interface{}); ok {
			path := &pbxrpl.Path{
				Elements: make([]*pbxrpl.PathElement, 0, len(pathArr)),
			}
			for _, elemRaw := range pathArr {
				if elemMap, ok := elemRaw.(map[string]interface{}); ok {
					elem := &pbxrpl.PathElement{}
					if account, ok := elemMap["account"].(string); ok {
						elem.Account = account
					}
					if currency, ok := elemMap["currency"].(string); ok {
						elem.Currency = currency
					}
					if issuer, ok := elemMap["issuer"].(string); ok {
						elem.Issuer = issuer
					}
					path.Elements = append(path.Elements, elem)
				}
			}
			result = append(result, path)
		}
	}
	return result
}

func (m *Mapper) mapAssetFromFlat(assetRaw interface{}) *pbxrpl.Asset {
	if assetRaw == nil {
		return nil
	}

	if assetMap, ok := assetRaw.(map[string]interface{}); ok {
		asset := &pbxrpl.Asset{}
		if currency, ok := assetMap["currency"].(string); ok {
			asset.Currency = currency
		}
		if issuer, ok := assetMap["issuer"].(string); ok {
			asset.Issuer = issuer
		}
		if mptID, ok := assetMap["mpt_issuance_id"].(string); ok {
			asset.MptIssuanceId = mptID
		}
		return asset
	}

	return nil
}

func (m *Mapper) mapSignerEntries(entriesRaw []interface{}) []*pbxrpl.SignerEntry {
	result := make([]*pbxrpl.SignerEntry, 0, len(entriesRaw))
	for _, entryRaw := range entriesRaw {
		if entryWrapper, ok := entryRaw.(map[string]interface{}); ok {
			if entry, ok := entryWrapper["SignerEntry"].(map[string]interface{}); ok {
				se := &pbxrpl.SignerEntry{}
				if account, ok := entry["Account"].(string); ok {
					se.Account = account
				}
				if weight, ok := asUint32(entry["SignerWeight"]); ok {
					se.SignerWeight = weight
				}
				if walletLocator, ok := entry["WalletLocator"].(string); ok {
					se.WalletLocator = walletLocator
				}
				result = append(result, se)
			}
		}
	}
	return result
}

func (m *Mapper) mapAuthAccounts(accountsRaw []interface{}) []*pbxrpl.AuthAccount {
	result := make([]*pbxrpl.AuthAccount, 0, len(accountsRaw))
	for _, accountRaw := range accountsRaw {
		if accountWrapper, ok := accountRaw.(map[string]interface{}); ok {
			if account, ok := accountWrapper["AuthAccount"].(map[string]interface{}); ok {
				aa := &pbxrpl.AuthAccount{}
				if acct, ok := account["Account"].(string); ok {
					aa.Account = acct
				}
				result = append(result, aa)
			}
		}
	}
	return result
}

func (m *Mapper) mapPriceDataSeries(seriesRaw []interface{}) []*pbxrpl.PriceData {
	result := make([]*pbxrpl.PriceData, 0, len(seriesRaw))
	for _, dataRaw := range seriesRaw {
		if dataWrapper, ok := dataRaw.(map[string]interface{}); ok {
			if data, ok := dataWrapper["PriceData"].(map[string]interface{}); ok {
				pd := &pbxrpl.PriceData{}
				if baseAsset, ok := data["BaseAsset"].(string); ok {
					pd.BaseAsset = baseAsset
				}
				if quoteAsset, ok := data["QuoteAsset"].(string); ok {
					pd.QuoteAsset = quoteAsset
				}
				if assetPrice, ok := data["AssetPrice"].(string); ok {
					if parsed, err := strconv.ParseUint(assetPrice, 10, 64); err == nil {
						pd.AssetPrice = parsed
					}
				}
				if scale, ok := asUint32(data["Scale"]); ok {
					pd.Scale = scale
				}
				result = append(result, pd)
			}
		}
	}
	return result
}

func (m *Mapper) mapRawTransactions(txsRaw []interface{}) []*pbxrpl.RawTransaction {
	result := make([]*pbxrpl.RawTransaction, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		if txMap, ok := txRaw.(map[string]interface{}); ok {
			rt := &pbxrpl.RawTransaction{}
			// The raw transaction would be bytes/hex encoded
			if rawTx, ok := txMap["RawTransaction"].(string); ok {
				// Convert hex string to bytes
				rt.RawTransaction = []byte(rawTx)
			}
			result = append(result, rt)
		}
	}
	return result
}

func (m *Mapper) mapBatchSigners(signersRaw []interface{}) []*pbxrpl.BatchSigner {
	result := make([]*pbxrpl.BatchSigner, 0, len(signersRaw))
	for _, signerRaw := range signersRaw {
		if signerMap, ok := signerRaw.(map[string]interface{}); ok {
			bs := &pbxrpl.BatchSigner{}
			if account, ok := signerMap["Account"].(string); ok {
				bs.Account = account
			}
			if signingPubKey, ok := signerMap["SigningPubKey"].(string); ok {
				bs.SigningPubKey = signingPubKey
			}
			if txnSignature, ok := signerMap["TxnSignature"].(string); ok {
				bs.TxnSignature = txnSignature
			}
			if signers, ok := signerMap["Signers"].([]interface{}); ok {
				bs.Signers = m.mapSignersFromFlat(signers)
			}
			result = append(result, bs)
		}
	}
	return result
}
