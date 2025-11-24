package types

import (
	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
)

// TransactionMeta holds decoded transaction data for internal processing
type TransactionMeta struct {
	Hash      []byte
	TxBlob    []byte
	MetaBlob  []byte
	TxType    pbxrpl.TransactionType
	Result    pbxrpl.TransactionResult
	Account   string
	Fee       uint64
	Sequence  uint32
}

func NewTransactionMeta(
	hash []byte,
	txBlob []byte,
	metaBlob []byte,
	txType pbxrpl.TransactionType,
	result pbxrpl.TransactionResult,
	account string,
	fee uint64,
	sequence uint32,
) *TransactionMeta {
	return &TransactionMeta{
		Hash:     hash,
		TxBlob:   txBlob,
		MetaBlob: metaBlob,
		TxType:   txType,
		Result:   result,
		Account:  account,
		Fee:      fee,
		Sequence: sequence,
	}
}

// TxRequest represents a request to fetch a specific transaction
type TxRequest struct {
	Method string     `json:"method"`
	Params []TxParams `json:"params"`
}

type TxParams struct {
	Transaction string `json:"transaction"`
	Binary      bool   `json:"binary,omitempty"`
}

func NewTxRequest(txHash string, binary bool) *TxRequest {
	return &TxRequest{
		Method: "tx",
		Params: []TxParams{{
			Transaction: txHash,
			Binary:      binary,
		}},
	}
}

// TxResponse represents the response from the tx method
type TxResponse struct {
	Result TxResult `json:"result"`
}

type TxResult struct {
	Hash            string `json:"hash"`
	LedgerIndex     uint64 `json:"ledger_index"`
	Status          string `json:"status"`
	Validated       bool   `json:"validated"`
	Meta            any    `json:"meta,omitempty"`        // JSON or binary
	MetaBlob        string `json:"meta_blob,omitempty"`   // When binary=true
	TxBlob          string `json:"tx_blob,omitempty"`     // When binary=true
	TransactionType string `json:"TransactionType,omitempty"`
	Account         string `json:"Account,omitempty"`
	Fee             string `json:"Fee,omitempty"`
	Sequence        uint32 `json:"Sequence,omitempty"`
}

// XRPL Transaction Type codes (from rippled source)
// These match the binary encoding values
const (
	TxTypePayment              uint16 = 0
	TxTypeEscrowCreate         uint16 = 1
	TxTypeEscrowFinish         uint16 = 2
	TxTypeAccountSet           uint16 = 3
	TxTypeEscrowCancel         uint16 = 4
	TxTypeRegularKeySet        uint16 = 5
	TxTypeNickNameSet          uint16 = 6  // Deprecated
	TxTypeOfferCreate          uint16 = 7
	TxTypeOfferCancel          uint16 = 8
	TxTypeContract             uint16 = 9  // Deprecated
	TxTypeTicketCreate         uint16 = 10
	TxTypeTicketCancel         uint16 = 11 // Deprecated
	TxTypeSignerListSet        uint16 = 12
	TxTypePaymentChannelCreate uint16 = 13
	TxTypePaymentChannelFund   uint16 = 14
	TxTypePaymentChannelClaim  uint16 = 15
	TxTypeCheckCreate          uint16 = 16
	TxTypeCheckCash            uint16 = 17
	TxTypeCheckCancel          uint16 = 18
	TxTypeDepositPreauth       uint16 = 19
	TxTypeTrustSet             uint16 = 20
	TxTypeAccountDelete        uint16 = 21
	TxTypeSetHook              uint16 = 22
	TxTypeNFTokenMint          uint16 = 25
	TxTypeNFTokenBurn          uint16 = 26
	TxTypeNFTokenCreateOffer   uint16 = 27
	TxTypeNFTokenCancelOffer   uint16 = 28
	TxTypeNFTokenAcceptOffer   uint16 = 29
	TxTypeClawback             uint16 = 30
	TxTypeAMMCreate            uint16 = 35
	TxTypeAMMDeposit           uint16 = 36
	TxTypeAMMWithdraw          uint16 = 37
	TxTypeAMMVote              uint16 = 38
	TxTypeAMMBid               uint16 = 39
	TxTypeAMMDelete            uint16 = 40
	TxTypeXChainCreateBridge   uint16 = 41
	TxTypeXChainCreateClaimID  uint16 = 42
	TxTypeXChainCommit         uint16 = 43
	TxTypeXChainClaim          uint16 = 44
	TxTypeXChainAccountCreate  uint16 = 45
	TxTypeXChainAddClaimAttestation uint16 = 46
	TxTypeXChainAddAccountCreateAttestation uint16 = 47
	TxTypeXChainModifyBridge   uint16 = 48
	TxTypeDIDSet               uint16 = 49
	TxTypeDIDDelete            uint16 = 50
	TxTypeOracleSet            uint16 = 51
	TxTypeOracleDelete         uint16 = 52
)

// TxTypeCodeToProto converts XRPL binary tx type code to protobuf enum
func TxTypeCodeToProto(code uint16) pbxrpl.TransactionType {
	switch code {
	case TxTypePayment:
		return pbxrpl.TransactionType_TX_PAYMENT
	case TxTypeEscrowCreate:
		return pbxrpl.TransactionType_TX_ESCROW_CREATE
	case TxTypeEscrowFinish:
		return pbxrpl.TransactionType_TX_ESCROW_FINISH
	case TxTypeAccountSet:
		return pbxrpl.TransactionType_TX_ACCOUNT_SET
	case TxTypeEscrowCancel:
		return pbxrpl.TransactionType_TX_ESCROW_CANCEL
	case TxTypeRegularKeySet:
		return pbxrpl.TransactionType_TX_SET_REGULAR_KEY
	case TxTypeOfferCreate:
		return pbxrpl.TransactionType_TX_OFFER_CREATE
	case TxTypeOfferCancel:
		return pbxrpl.TransactionType_TX_OFFER_CANCEL
	case TxTypeTicketCreate:
		return pbxrpl.TransactionType_TX_TICKET_CREATE
	case TxTypeSignerListSet:
		return pbxrpl.TransactionType_TX_SIGNER_LIST_SET
	case TxTypePaymentChannelCreate:
		return pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_CREATE
	case TxTypePaymentChannelFund:
		return pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_FUND
	case TxTypePaymentChannelClaim:
		return pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_CLAIM
	case TxTypeCheckCreate:
		return pbxrpl.TransactionType_TX_CHECK_CREATE
	case TxTypeCheckCash:
		return pbxrpl.TransactionType_TX_CHECK_CASH
	case TxTypeCheckCancel:
		return pbxrpl.TransactionType_TX_CHECK_CANCEL
	case TxTypeDepositPreauth:
		return pbxrpl.TransactionType_TX_DEPOSIT_PREAUTH
	case TxTypeTrustSet:
		return pbxrpl.TransactionType_TX_TRUST_SET
	case TxTypeAccountDelete:
		return pbxrpl.TransactionType_TX_ACCOUNT_DELETE
	case TxTypeNFTokenMint:
		return pbxrpl.TransactionType_TX_NFT_MINT
	case TxTypeNFTokenBurn:
		return pbxrpl.TransactionType_TX_NFT_BURN
	case TxTypeNFTokenCreateOffer:
		return pbxrpl.TransactionType_TX_NFT_CREATE_OFFER
	case TxTypeNFTokenCancelOffer:
		return pbxrpl.TransactionType_TX_NFT_CANCEL_OFFER
	case TxTypeNFTokenAcceptOffer:
		return pbxrpl.TransactionType_TX_NFT_ACCEPT_OFFER
	case TxTypeClawback:
		return pbxrpl.TransactionType_TX_CLAWBACK
	case TxTypeAMMCreate:
		return pbxrpl.TransactionType_TX_AMM_CREATE
	case TxTypeAMMDeposit:
		return pbxrpl.TransactionType_TX_AMM_DEPOSIT
	case TxTypeAMMWithdraw:
		return pbxrpl.TransactionType_TX_AMM_WITHDRAW
	case TxTypeAMMVote:
		return pbxrpl.TransactionType_TX_AMM_VOTE
	case TxTypeAMMBid:
		return pbxrpl.TransactionType_TX_AMM_BID
	case TxTypeAMMDelete:
		return pbxrpl.TransactionType_TX_AMM_DELETE
	case TxTypeXChainCreateBridge:
		return pbxrpl.TransactionType_TX_XCHAIN_CREATE_BRIDGE
	case TxTypeXChainCreateClaimID:
		return pbxrpl.TransactionType_TX_XCHAIN_CREATE_CLAIM_ID
	case TxTypeXChainCommit:
		return pbxrpl.TransactionType_TX_XCHAIN_COMMIT
	case TxTypeXChainClaim:
		return pbxrpl.TransactionType_TX_XCHAIN_CLAIM
	case TxTypeXChainAccountCreate:
		return pbxrpl.TransactionType_TX_XCHAIN_ACCOUNT_CREATE_COMMIT
	case TxTypeXChainAddClaimAttestation:
		return pbxrpl.TransactionType_TX_XCHAIN_ADD_CLAIM_ATTESTATION
	case TxTypeXChainAddAccountCreateAttestation:
		return pbxrpl.TransactionType_TX_XCHAIN_ADD_ACCOUNT_CREATE_ATTESTATION
	case TxTypeXChainModifyBridge:
		return pbxrpl.TransactionType_TX_XCHAIN_MODIFY_BRIDGE
	case TxTypeDIDSet:
		return pbxrpl.TransactionType_TX_DID_SET
	case TxTypeDIDDelete:
		return pbxrpl.TransactionType_TX_DID_DELETE
	case TxTypeOracleSet:
		return pbxrpl.TransactionType_TX_ORACLE_SET
	case TxTypeOracleDelete:
		return pbxrpl.TransactionType_TX_ORACLE_DELETE
	default:
		return pbxrpl.TransactionType_TX_UNKNOWN
	}
}

// TxTypeStringToProto converts transaction type string to protobuf enum
func TxTypeStringToProto(txType string) pbxrpl.TransactionType {
	switch txType {
	case "Payment":
		return pbxrpl.TransactionType_TX_PAYMENT
	case "EscrowCreate":
		return pbxrpl.TransactionType_TX_ESCROW_CREATE
	case "EscrowFinish":
		return pbxrpl.TransactionType_TX_ESCROW_FINISH
	case "AccountSet":
		return pbxrpl.TransactionType_TX_ACCOUNT_SET
	case "EscrowCancel":
		return pbxrpl.TransactionType_TX_ESCROW_CANCEL
	case "SetRegularKey":
		return pbxrpl.TransactionType_TX_SET_REGULAR_KEY
	case "OfferCreate":
		return pbxrpl.TransactionType_TX_OFFER_CREATE
	case "OfferCancel":
		return pbxrpl.TransactionType_TX_OFFER_CANCEL
	case "TicketCreate":
		return pbxrpl.TransactionType_TX_TICKET_CREATE
	case "SignerListSet":
		return pbxrpl.TransactionType_TX_SIGNER_LIST_SET
	case "PaymentChannelCreate":
		return pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_CREATE
	case "PaymentChannelFund":
		return pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_FUND
	case "PaymentChannelClaim":
		return pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_CLAIM
	case "CheckCreate":
		return pbxrpl.TransactionType_TX_CHECK_CREATE
	case "CheckCash":
		return pbxrpl.TransactionType_TX_CHECK_CASH
	case "CheckCancel":
		return pbxrpl.TransactionType_TX_CHECK_CANCEL
	case "DepositPreauth":
		return pbxrpl.TransactionType_TX_DEPOSIT_PREAUTH
	case "TrustSet":
		return pbxrpl.TransactionType_TX_TRUST_SET
	case "AccountDelete":
		return pbxrpl.TransactionType_TX_ACCOUNT_DELETE
	case "NFTokenMint":
		return pbxrpl.TransactionType_TX_NFT_MINT
	case "NFTokenBurn":
		return pbxrpl.TransactionType_TX_NFT_BURN
	case "NFTokenCreateOffer":
		return pbxrpl.TransactionType_TX_NFT_CREATE_OFFER
	case "NFTokenCancelOffer":
		return pbxrpl.TransactionType_TX_NFT_CANCEL_OFFER
	case "NFTokenAcceptOffer":
		return pbxrpl.TransactionType_TX_NFT_ACCEPT_OFFER
	case "Clawback":
		return pbxrpl.TransactionType_TX_CLAWBACK
	case "AMMCreate":
		return pbxrpl.TransactionType_TX_AMM_CREATE
	case "AMMDeposit":
		return pbxrpl.TransactionType_TX_AMM_DEPOSIT
	case "AMMWithdraw":
		return pbxrpl.TransactionType_TX_AMM_WITHDRAW
	case "AMMVote":
		return pbxrpl.TransactionType_TX_AMM_VOTE
	case "AMMBid":
		return pbxrpl.TransactionType_TX_AMM_BID
	case "AMMDelete":
		return pbxrpl.TransactionType_TX_AMM_DELETE
	case "XChainCreateBridge":
		return pbxrpl.TransactionType_TX_XCHAIN_CREATE_BRIDGE
	case "XChainCreateClaimID":
		return pbxrpl.TransactionType_TX_XCHAIN_CREATE_CLAIM_ID
	case "XChainCommit":
		return pbxrpl.TransactionType_TX_XCHAIN_COMMIT
	case "XChainClaim":
		return pbxrpl.TransactionType_TX_XCHAIN_CLAIM
	case "XChainAccountCreateCommit":
		return pbxrpl.TransactionType_TX_XCHAIN_ACCOUNT_CREATE_COMMIT
	case "XChainAddClaimAttestation":
		return pbxrpl.TransactionType_TX_XCHAIN_ADD_CLAIM_ATTESTATION
	case "XChainAddAccountCreateAttestation":
		return pbxrpl.TransactionType_TX_XCHAIN_ADD_ACCOUNT_CREATE_ATTESTATION
	case "XChainModifyBridge":
		return pbxrpl.TransactionType_TX_XCHAIN_MODIFY_BRIDGE
	case "DIDSet":
		return pbxrpl.TransactionType_TX_DID_SET
	case "DIDDelete":
		return pbxrpl.TransactionType_TX_DID_DELETE
	case "OracleSet":
		return pbxrpl.TransactionType_TX_ORACLE_SET
	case "OracleDelete":
		return pbxrpl.TransactionType_TX_ORACLE_DELETE
	default:
		return pbxrpl.TransactionType_TX_UNKNOWN
	}
}
