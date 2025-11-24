package decoder

import (
	"strconv"

	pbxrpl "github.com/xrpl-commons/firehose-xrpl/pb/sf/xrpl/type/v1"
)

// DecodeTransactionDetails converts the decoded JSON map to the appropriate protobuf tx_details
// Returns the interface that should be assigned to the Transaction's oneof field
func (d *Decoder) DecodeTransactionDetails(decoded map[string]interface{}, txType pbxrpl.TransactionType) interface{} {
	switch txType {
	case pbxrpl.TransactionType_TX_PAYMENT:
		return d.decodePayment(decoded)
	case pbxrpl.TransactionType_TX_OFFER_CREATE:
		return d.decodeOfferCreate(decoded)
	case pbxrpl.TransactionType_TX_OFFER_CANCEL:
		return d.decodeOfferCancel(decoded)
	case pbxrpl.TransactionType_TX_TRUST_SET:
		return d.decodeTrustSet(decoded)
	case pbxrpl.TransactionType_TX_ACCOUNT_SET:
		return d.decodeAccountSet(decoded)
	case pbxrpl.TransactionType_TX_ACCOUNT_DELETE:
		return d.decodeAccountDelete(decoded)
	case pbxrpl.TransactionType_TX_SET_REGULAR_KEY:
		return d.decodeSetRegularKey(decoded)
	case pbxrpl.TransactionType_TX_SIGNER_LIST_SET:
		return d.decodeSignerListSet(decoded)
	case pbxrpl.TransactionType_TX_ESCROW_CREATE:
		return d.decodeEscrowCreate(decoded)
	case pbxrpl.TransactionType_TX_ESCROW_FINISH:
		return d.decodeEscrowFinish(decoded)
	case pbxrpl.TransactionType_TX_ESCROW_CANCEL:
		return d.decodeEscrowCancel(decoded)
	case pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_CREATE:
		return d.decodePaymentChannelCreate(decoded)
	case pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_FUND:
		return d.decodePaymentChannelFund(decoded)
	case pbxrpl.TransactionType_TX_PAYMENT_CHANNEL_CLAIM:
		return d.decodePaymentChannelClaim(decoded)
	case pbxrpl.TransactionType_TX_CHECK_CREATE:
		return d.decodeCheckCreate(decoded)
	case pbxrpl.TransactionType_TX_CHECK_CASH:
		return d.decodeCheckCash(decoded)
	case pbxrpl.TransactionType_TX_CHECK_CANCEL:
		return d.decodeCheckCancel(decoded)
	case pbxrpl.TransactionType_TX_DEPOSIT_PREAUTH:
		return d.decodeDepositPreauth(decoded)
	case pbxrpl.TransactionType_TX_TICKET_CREATE:
		return d.decodeTicketCreate(decoded)
	case pbxrpl.TransactionType_TX_NFT_MINT:
		return d.decodeNFTokenMint(decoded)
	case pbxrpl.TransactionType_TX_NFT_BURN:
		return d.decodeNFTokenBurn(decoded)
	case pbxrpl.TransactionType_TX_NFT_CREATE_OFFER:
		return d.decodeNFTokenCreateOffer(decoded)
	case pbxrpl.TransactionType_TX_NFT_CANCEL_OFFER:
		return d.decodeNFTokenCancelOffer(decoded)
	case pbxrpl.TransactionType_TX_NFT_ACCEPT_OFFER:
		return d.decodeNFTokenAcceptOffer(decoded)
	case pbxrpl.TransactionType_TX_CLAWBACK:
		return d.decodeClawback(decoded)
	case pbxrpl.TransactionType_TX_AMM_CREATE:
		return d.decodeAMMCreate(decoded)
	case pbxrpl.TransactionType_TX_AMM_DEPOSIT:
		return d.decodeAMMDeposit(decoded)
	case pbxrpl.TransactionType_TX_AMM_WITHDRAW:
		return d.decodeAMMWithdraw(decoded)
	case pbxrpl.TransactionType_TX_AMM_VOTE:
		return d.decodeAMMVote(decoded)
	case pbxrpl.TransactionType_TX_AMM_BID:
		return d.decodeAMMBid(decoded)
	case pbxrpl.TransactionType_TX_AMM_DELETE:
		return d.decodeAMMDelete(decoded)
	default:
		return nil
	}
}

// Helper functions for extracting values from decoded JSON

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getUint32(m map[string]interface{}, key string) uint32 {
	switch v := m[key].(type) {
	case float64:
		return uint32(v)
	case int:
		return uint32(v)
	case int64:
		return uint32(v)
	}
	return 0
}

func getUint64(m map[string]interface{}, key string) uint64 {
	switch v := m[key].(type) {
	case float64:
		return uint64(v)
	case int:
		return uint64(v)
	case int64:
		return uint64(v)
	case string:
		if val, err := strconv.ParseUint(v, 10, 64); err == nil {
			return val
		}
	}
	return 0
}

func getBool(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	// XRPL uses 1/0 for booleans in some cases
	switch v := m[key].(type) {
	case float64:
		return v == 1
	case int:
		return v == 1
	}
	return false
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if arr, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// decodeAmount converts XRPL amount format to protobuf Amount
// XRPL amounts can be:
// - string: XRP drops (e.g., "1000000" = 1 XRP)
// - object: IOU with value, currency, issuer
func decodeAmount(v interface{}) *pbxrpl.Amount {
	if v == nil {
		return nil
	}

	switch amt := v.(type) {
	case string:
		// XRP amount in drops
		return &pbxrpl.Amount{
			Value: amt,
			// Currency and Issuer empty for XRP
		}
	case map[string]interface{}:
		// IOU amount
		return &pbxrpl.Amount{
			Value:    getString(amt, "value"),
			Currency: getString(amt, "currency"),
			Issuer:   getString(amt, "issuer"),
		}
	}
	return nil
}

// decodeAsset converts XRPL asset/issue format to protobuf Asset
func decodeAsset(v interface{}) *pbxrpl.Asset {
	if v == nil {
		return nil
	}

	if m, ok := v.(map[string]interface{}); ok {
		return &pbxrpl.Asset{
			Currency: getString(m, "currency"),
			Issuer:   getString(m, "issuer"),
		}
	}
	return nil
}

// decodePaths converts XRPL paths array to protobuf Paths
func decodePaths(v interface{}) []*pbxrpl.Path {
	if v == nil {
		return nil
	}

	paths, ok := v.([]interface{})
	if !ok {
		return nil
	}

	result := make([]*pbxrpl.Path, 0, len(paths))
	for _, p := range paths {
		pathArr, ok := p.([]interface{})
		if !ok {
			continue
		}

		path := &pbxrpl.Path{
			Elements: make([]*pbxrpl.PathElement, 0, len(pathArr)),
		}

		for _, elem := range pathArr {
			elemMap, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}

			path.Elements = append(path.Elements, &pbxrpl.PathElement{
				Account:  getString(elemMap, "account"),
				Currency: getString(elemMap, "currency"),
				Issuer:   getString(elemMap, "issuer"),
			})
		}

		result = append(result, path)
	}

	return result
}

// Transaction type decoders

func (d *Decoder) decodePayment(m map[string]interface{}) *pbxrpl.Payment {
	return &pbxrpl.Payment{
		Destination:    getString(m, "Destination"),
		Amount:         decodeAmount(m["Amount"]),
		SendMax:        decodeAmount(m["SendMax"]),
		DeliverMin:     decodeAmount(m["DeliverMin"]),
		Paths:          decodePaths(m["Paths"]),
		InvoiceId:      getString(m, "InvoiceID"),
		DestinationTag: getUint32(m, "DestinationTag"),
	}
}

func (d *Decoder) decodeOfferCreate(m map[string]interface{}) *pbxrpl.OfferCreate {
	return &pbxrpl.OfferCreate{
		TakerGets:     decodeAmount(m["TakerGets"]),
		TakerPays:     decodeAmount(m["TakerPays"]),
		Expiration:    getUint32(m, "Expiration"),
		OfferSequence: getUint32(m, "OfferSequence"),
	}
}

func (d *Decoder) decodeOfferCancel(m map[string]interface{}) *pbxrpl.OfferCancel {
	return &pbxrpl.OfferCancel{
		OfferSequence: getUint32(m, "OfferSequence"),
	}
}

func (d *Decoder) decodeTrustSet(m map[string]interface{}) *pbxrpl.TrustSet {
	return &pbxrpl.TrustSet{
		LimitAmount: decodeAmount(m["LimitAmount"]),
		QualityIn:   getUint32(m, "QualityIn"),
		QualityOut:  getUint32(m, "QualityOut"),
	}
}

func (d *Decoder) decodeAccountSet(m map[string]interface{}) *pbxrpl.AccountSet {
	return &pbxrpl.AccountSet{
		SetFlag:       getUint32(m, "SetFlag"),
		ClearFlag:     getUint32(m, "ClearFlag"),
		Domain:        getString(m, "Domain"),
		EmailHash:     getString(m, "EmailHash"),
		MessageKey:    getString(m, "MessageKey"),
		TransferRate:  getUint32(m, "TransferRate"),
		TickSize:      getUint32(m, "TickSize"),
		NftokenMinter: getString(m, "NFTokenMinter"),
	}
}

func (d *Decoder) decodeAccountDelete(m map[string]interface{}) *pbxrpl.AccountDelete {
	return &pbxrpl.AccountDelete{
		Destination:    getString(m, "Destination"),
		DestinationTag: getUint32(m, "DestinationTag"),
	}
}

func (d *Decoder) decodeSetRegularKey(m map[string]interface{}) *pbxrpl.SetRegularKey {
	return &pbxrpl.SetRegularKey{
		RegularKey: getString(m, "RegularKey"),
	}
}

func (d *Decoder) decodeSignerListSet(m map[string]interface{}) *pbxrpl.SignerListSet {
	result := &pbxrpl.SignerListSet{
		SignerQuorum: getUint32(m, "SignerQuorum"),
	}

	if entries, ok := m["SignerEntries"].([]interface{}); ok {
		result.SignerEntries = make([]*pbxrpl.SignerEntry, 0, len(entries))
		for _, e := range entries {
			if entryWrapper, ok := e.(map[string]interface{}); ok {
				if entry, ok := entryWrapper["SignerEntry"].(map[string]interface{}); ok {
					result.SignerEntries = append(result.SignerEntries, &pbxrpl.SignerEntry{
						Account:      getString(entry, "Account"),
						SignerWeight: getUint32(entry, "SignerWeight"),
					})
				}
			}
		}
	}

	return result
}

func (d *Decoder) decodeEscrowCreate(m map[string]interface{}) *pbxrpl.EscrowCreate {
	return &pbxrpl.EscrowCreate{
		Amount:         decodeAmount(m["Amount"]),
		Destination:    getString(m, "Destination"),
		CancelAfter:    getUint32(m, "CancelAfter"),
		FinishAfter:    getUint32(m, "FinishAfter"),
		Condition:      getString(m, "Condition"),
		DestinationTag: getUint32(m, "DestinationTag"),
	}
}

func (d *Decoder) decodeEscrowFinish(m map[string]interface{}) *pbxrpl.EscrowFinish {
	return &pbxrpl.EscrowFinish{
		Owner:         getString(m, "Owner"),
		OfferSequence: getUint32(m, "OfferSequence"),
		Condition:     getString(m, "Condition"),
		Fulfillment:   getString(m, "Fulfillment"),
	}
}

func (d *Decoder) decodeEscrowCancel(m map[string]interface{}) *pbxrpl.EscrowCancel {
	return &pbxrpl.EscrowCancel{
		Owner:         getString(m, "Owner"),
		OfferSequence: getUint32(m, "OfferSequence"),
	}
}

func (d *Decoder) decodePaymentChannelCreate(m map[string]interface{}) *pbxrpl.PaymentChannelCreate {
	return &pbxrpl.PaymentChannelCreate{
		Destination:    getString(m, "Destination"),
		Amount:         decodeAmount(m["Amount"]),
		SettleDelay:    getUint32(m, "SettleDelay"),
		PublicKey:      getString(m, "PublicKey"),
		CancelAfter:    getUint32(m, "CancelAfter"),
		DestinationTag: getUint32(m, "DestinationTag"),
	}
}

func (d *Decoder) decodePaymentChannelFund(m map[string]interface{}) *pbxrpl.PaymentChannelFund {
	return &pbxrpl.PaymentChannelFund{
		Channel:    getString(m, "Channel"),
		Amount:     decodeAmount(m["Amount"]),
		Expiration: getUint32(m, "Expiration"),
	}
}

func (d *Decoder) decodePaymentChannelClaim(m map[string]interface{}) *pbxrpl.PaymentChannelClaim {
	return &pbxrpl.PaymentChannelClaim{
		Channel:   getString(m, "Channel"),
		Amount:    decodeAmount(m["Amount"]),
		Balance:   decodeAmount(m["Balance"]),
		Signature: getString(m, "Signature"),
		PublicKey: getString(m, "PublicKey"),
	}
}

func (d *Decoder) decodeCheckCreate(m map[string]interface{}) *pbxrpl.CheckCreate {
	return &pbxrpl.CheckCreate{
		Destination:    getString(m, "Destination"),
		SendMax:        decodeAmount(m["SendMax"]),
		Expiration:     getUint32(m, "Expiration"),
		DestinationTag: getUint32(m, "DestinationTag"),
		InvoiceId:      getString(m, "InvoiceID"),
	}
}

func (d *Decoder) decodeCheckCash(m map[string]interface{}) *pbxrpl.CheckCash {
	return &pbxrpl.CheckCash{
		CheckId:    getString(m, "CheckID"),
		Amount:     decodeAmount(m["Amount"]),
		DeliverMin: decodeAmount(m["DeliverMin"]),
	}
}

func (d *Decoder) decodeCheckCancel(m map[string]interface{}) *pbxrpl.CheckCancel {
	return &pbxrpl.CheckCancel{
		CheckId: getString(m, "CheckID"),
	}
}

func (d *Decoder) decodeDepositPreauth(m map[string]interface{}) *pbxrpl.DepositPreauth {
	return &pbxrpl.DepositPreauth{
		Authorize:   getString(m, "Authorize"),
		Unauthorize: getString(m, "Unauthorize"),
		// AuthorizeCredentials and UnauthorizeCredentials would need additional parsing
	}
}

func (d *Decoder) decodeTicketCreate(m map[string]interface{}) *pbxrpl.TicketCreate {
	return &pbxrpl.TicketCreate{
		TicketCount: getUint32(m, "TicketCount"),
	}
}

func (d *Decoder) decodeNFTokenMint(m map[string]interface{}) *pbxrpl.NFTokenMint {
	return &pbxrpl.NFTokenMint{
		NftokenTaxon: getUint32(m, "NFTokenTaxon"),
		Issuer:       getString(m, "Issuer"),
		TransferFee:  getUint32(m, "TransferFee"),
		Uri:          getString(m, "URI"),
		Flags:        getUint32(m, "Flags"),
	}
}

func (d *Decoder) decodeNFTokenBurn(m map[string]interface{}) *pbxrpl.NFTokenBurn {
	return &pbxrpl.NFTokenBurn{
		NftokenId: getString(m, "NFTokenID"),
		Owner:     getString(m, "Owner"),
	}
}

func (d *Decoder) decodeNFTokenCreateOffer(m map[string]interface{}) *pbxrpl.NFTokenCreateOffer {
	return &pbxrpl.NFTokenCreateOffer{
		NftokenId:   getString(m, "NFTokenID"),
		Amount:      decodeAmount(m["Amount"]),
		Owner:       getString(m, "Owner"),
		Destination: getString(m, "Destination"),
		Expiration:  getUint32(m, "Expiration"),
		Flags:       getUint32(m, "Flags"),
	}
}

func (d *Decoder) decodeNFTokenCancelOffer(m map[string]interface{}) *pbxrpl.NFTokenCancelOffer {
	return &pbxrpl.NFTokenCancelOffer{
		NftokenOffers: getStringSlice(m, "NFTokenOffers"),
	}
}

func (d *Decoder) decodeNFTokenAcceptOffer(m map[string]interface{}) *pbxrpl.NFTokenAcceptOffer {
	return &pbxrpl.NFTokenAcceptOffer{
		NftokenSellOffer:  getString(m, "NFTokenSellOffer"),
		NftokenBuyOffer:   getString(m, "NFTokenBuyOffer"),
		NftokenBrokerFee:  decodeAmount(m["NFTokenBrokerFee"]),
	}
}

func (d *Decoder) decodeClawback(m map[string]interface{}) *pbxrpl.Clawback {
	return &pbxrpl.Clawback{
		Amount: decodeAmount(m["Amount"]),
		Holder: getString(m, "Holder"),
	}
}

func (d *Decoder) decodeAMMCreate(m map[string]interface{}) *pbxrpl.AMMCreate {
	return &pbxrpl.AMMCreate{
		Amount:     decodeAmount(m["Amount"]),
		Amount2:    decodeAmount(m["Amount2"]),
		TradingFee: getUint32(m, "TradingFee"),
	}
}

func (d *Decoder) decodeAMMDeposit(m map[string]interface{}) *pbxrpl.AMMDeposit {
	return &pbxrpl.AMMDeposit{
		Asset:      decodeAsset(m["Asset"]),
		Asset2:     decodeAsset(m["Asset2"]),
		Amount:     decodeAmount(m["Amount"]),
		Amount2:    decodeAmount(m["Amount2"]),
		EPrice:     decodeAmount(m["EPrice"]),
		LpTokenOut: decodeAmount(m["LPTokenOut"]),
		TradingFee: getUint32(m, "TradingFee"),
	}
}

func (d *Decoder) decodeAMMWithdraw(m map[string]interface{}) *pbxrpl.AMMWithdraw {
	return &pbxrpl.AMMWithdraw{
		Asset:    decodeAsset(m["Asset"]),
		Asset2:   decodeAsset(m["Asset2"]),
		Amount:   decodeAmount(m["Amount"]),
		Amount2:  decodeAmount(m["Amount2"]),
		EPrice:   decodeAmount(m["EPrice"]),
		LpTokenIn: decodeAmount(m["LPTokenIn"]),
	}
}

func (d *Decoder) decodeAMMVote(m map[string]interface{}) *pbxrpl.AMMVote {
	return &pbxrpl.AMMVote{
		Asset:      decodeAsset(m["Asset"]),
		Asset2:     decodeAsset(m["Asset2"]),
		TradingFee: getUint32(m, "TradingFee"),
	}
}

func (d *Decoder) decodeAMMBid(m map[string]interface{}) *pbxrpl.AMMBid {
	result := &pbxrpl.AMMBid{
		Asset:  decodeAsset(m["Asset"]),
		Asset2: decodeAsset(m["Asset2"]),
		BidMin: decodeAmount(m["BidMin"]),
		BidMax: decodeAmount(m["BidMax"]),
	}

	// Parse AuthAccounts
	if authAccounts, ok := m["AuthAccounts"].([]interface{}); ok {
		result.AuthAccounts = make([]*pbxrpl.AuthAccount, 0, len(authAccounts))
		for _, aa := range authAccounts {
			if aaMap, ok := aa.(map[string]interface{}); ok {
				if authAccount, ok := aaMap["AuthAccount"].(map[string]interface{}); ok {
					result.AuthAccounts = append(result.AuthAccounts, &pbxrpl.AuthAccount{
						Account: getString(authAccount, "Account"),
					})
				}
			}
		}
	}

	return result
}

func (d *Decoder) decodeAMMDelete(m map[string]interface{}) *pbxrpl.AMMDelete {
	return &pbxrpl.AMMDelete{
		Asset:  decodeAsset(m["Asset"]),
		Asset2: decodeAsset(m["Asset2"]),
	}
}
