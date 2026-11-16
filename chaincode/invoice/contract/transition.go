package contract

import "fmt"

const (
	statusIssued    = "ISSUED"
	statusAccepted  = "ACCEPTED"
	statusPaid      = "PAID"
	statusCancelled = "CANCELLED"
)

const (
	actionAccept = "accept"
	actionPay    = "pay"
	actionCancel = "cancel"
)

// applyTransition is the invoice state machine.
// The caller MSP is checked here so the chaincode functions stay thin.
func applyTransition(status, action, callerMSP, issuerMSP, payerMSP string) (string, error) {
	switch action {
	case actionAccept:
		if callerMSP != payerMSP {
			return "", fmt.Errorf("only the payer organization %s can accept this invoice", payerMSP)
		}
		if status != statusIssued {
			return "", fmt.Errorf("cannot accept an invoice in status %s", status)
		}
		return statusAccepted, nil
	case actionPay:
		if callerMSP != payerMSP {
			return "", fmt.Errorf("only the payer organization %s can pay this invoice", payerMSP)
		}
		if status != statusAccepted {
			return "", fmt.Errorf("cannot pay an invoice in status %s", status)
		}
		return statusPaid, nil
	case actionCancel:
		if callerMSP != issuerMSP {
			return "", fmt.Errorf("only the issuer organization %s can cancel this invoice", issuerMSP)
		}
		if status != statusIssued && status != statusAccepted {
			return "", fmt.Errorf("cannot cancel an invoice in status %s", status)
		}
		return statusCancelled, nil
	default:
		return "", fmt.Errorf("unknown action %s", action)
	}
}
