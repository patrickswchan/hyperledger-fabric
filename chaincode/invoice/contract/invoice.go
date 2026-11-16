package contract

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// InvoiceContract is a two-organization invoice workflow.
// Org1 is a typical issuer and Org2 is a typical payer, but either
// organization may issue an invoice to the other.
type InvoiceContract struct {
	contractapi.Contract
}

// Invoice is the world-state record. Fields are alphabetical so JSON
// encoding is stable if another chaincode language writes the same keys.
type Invoice struct {
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	ID        string `json:"id"`
	IssuerMSP string `json:"issuerMsp"`
	Memo      string `json:"memo"`
	PayerMSP  string `json:"payerMsp"`
	Status    string `json:"status"`
}

// HistoryEntry is one update of an invoice key.
type HistoryEntry struct {
	IsDelete  bool     `json:"isDelete"`
	Timestamp string   `json:"timestamp"`
	TxID      string   `json:"txId"`
	Value     *Invoice `json:"value,omitempty"`
}

// Issue creates an invoice. The caller's MSP is stored as the issuer.
func (c *InvoiceContract) Issue(ctx contractapi.TransactionContextInterface, id string, payerMSP string, currency string, memo string, amount int64) (*Invoice, error) {
	if id == "" {
		return nil, fmt.Errorf("invoice id is required")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if currency == "" {
		return nil, fmt.Errorf("currency is required")
	}
	if payerMSP == "" {
		return nil, fmt.Errorf("payer MSP is required")
	}

	issuerMSP, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return nil, fmt.Errorf("read caller MSP: %w", err)
	}
	if payerMSP == issuerMSP {
		return nil, fmt.Errorf("payer organization must be different from the issuer")
	}

	key := invoiceKey(id)
	existing, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf("read invoice %s: %w", id, err)
	}
	if existing != nil {
		return nil, fmt.Errorf("invoice %s already exists", id)
	}

	invoice := Invoice{
		Amount:    amount,
		Currency:  currency,
		ID:        id,
		IssuerMSP: issuerMSP,
		Memo:      memo,
		PayerMSP:  payerMSP,
		Status:    statusIssued,
	}
	if err := putInvoice(ctx, &invoice); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// Accept moves ISSUED -> ACCEPTED. Only the payer organization can call it.
func (c *InvoiceContract) Accept(ctx contractapi.TransactionContextInterface, id string) (*Invoice, error) {
	return c.transition(ctx, id, actionAccept)
}

// Pay moves ACCEPTED -> PAID. Only the payer organization can call it.
func (c *InvoiceContract) Pay(ctx contractapi.TransactionContextInterface, id string) (*Invoice, error) {
	return c.transition(ctx, id, actionPay)
}

// Cancel moves ISSUED or ACCEPTED -> CANCELLED. Only the issuer can call it.
func (c *InvoiceContract) Cancel(ctx contractapi.TransactionContextInterface, id string) (*Invoice, error) {
	return c.transition(ctx, id, actionCancel)
}

// Read returns one invoice. This is a query: the client should evaluate it.
func (c *InvoiceContract) Read(ctx contractapi.TransactionContextInterface, id string) (*Invoice, error) {
	return readInvoice(ctx, id)
}

// List returns every invoice key in world state.
func (c *InvoiceContract) List(ctx contractapi.TransactionContextInterface) ([]*Invoice, error) {
	iterator, err := ctx.GetStub().GetStateByRange("invoice:", "invoice;")
	if err != nil {
		return nil, fmt.Errorf("list invoices: %w", err)
	}
	defer iterator.Close()

	invoices := make([]*Invoice, 0)
	for iterator.HasNext() {
		item, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("iterate invoices: %w", err)
		}
		var invoice Invoice
		if err := json.Unmarshal(item.Value, &invoice); err != nil {
			return nil, fmt.Errorf("decode %s: %w", item.Key, err)
		}
		invoices = append(invoices, &invoice)
	}
	return invoices, nil
}

// History returns every write to an invoice, newest first.
func (c *InvoiceContract) History(ctx contractapi.TransactionContextInterface, id string) ([]*HistoryEntry, error) {
	if _, err := readInvoice(ctx, id); err != nil {
		return nil, err
	}

	iterator, err := ctx.GetStub().GetHistoryForKey(invoiceKey(id))
	if err != nil {
		return nil, fmt.Errorf("read history for %s: %w", id, err)
	}
	defer iterator.Close()

	entries := make([]*HistoryEntry, 0)
	for iterator.HasNext() {
		mod, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("iterate history for %s: %w", id, err)
		}
		entry := &HistoryEntry{
			IsDelete: mod.IsDelete,
			TxID:     mod.TxId,
		}
		if mod.Timestamp != nil {
			entry.Timestamp = mod.Timestamp.AsTime().UTC().Format(time.RFC3339Nano)
		}
		if len(mod.Value) > 0 {
			var invoice Invoice
			if err := json.Unmarshal(mod.Value, &invoice); err != nil {
				return nil, fmt.Errorf("decode history value: %w", err)
			}
			entry.Value = &invoice
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (c *InvoiceContract) transition(ctx contractapi.TransactionContextInterface, id string, action string) (*Invoice, error) {
	invoice, err := readInvoice(ctx, id)
	if err != nil {
		return nil, err
	}
	callerMSP, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return nil, fmt.Errorf("read caller MSP: %w", err)
	}
	next, err := applyTransition(invoice.Status, action, callerMSP, invoice.IssuerMSP, invoice.PayerMSP)
	if err != nil {
		return nil, fmt.Errorf("invoice %s: %w", id, err)
	}
	invoice.Status = next
	if err := putInvoice(ctx, invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

func invoiceKey(id string) string {
	return "invoice:" + id
}

func readInvoice(ctx contractapi.TransactionContextInterface, id string) (*Invoice, error) {
	raw, err := ctx.GetStub().GetState(invoiceKey(id))
	if err != nil {
		return nil, fmt.Errorf("read invoice %s: %w", id, err)
	}
	if raw == nil {
		return nil, fmt.Errorf("invoice %s does not exist", id)
	}
	var invoice Invoice
	if err := json.Unmarshal(raw, &invoice); err != nil {
		return nil, fmt.Errorf("decode invoice %s: %w", id, err)
	}
	return &invoice, nil
}

func putInvoice(ctx contractapi.TransactionContextInterface, invoice *Invoice) error {
	raw, err := json.Marshal(invoice)
	if err != nil {
		return fmt.Errorf("encode invoice %s: %w", invoice.ID, err)
	}
	if err := ctx.GetStub().PutState(invoiceKey(invoice.ID), raw); err != nil {
		return fmt.Errorf("write invoice %s: %w", invoice.ID, err)
	}
	return nil
}
