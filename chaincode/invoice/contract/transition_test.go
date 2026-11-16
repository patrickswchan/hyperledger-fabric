package contract

import "testing"

func TestApplyTransition(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		action    string
		caller    string
		want      string
		wantError bool
	}{
		{name: "payer accepts issued", status: statusIssued, action: actionAccept, caller: "Org2MSP", want: statusAccepted},
		{name: "issuer cannot accept", status: statusIssued, action: actionAccept, caller: "Org1MSP", wantError: true},
		{name: "payer pays accepted", status: statusAccepted, action: actionPay, caller: "Org2MSP", want: statusPaid},
		{name: "cannot pay before accept", status: statusIssued, action: actionPay, caller: "Org2MSP", wantError: true},
		{name: "issuer cancels issued", status: statusIssued, action: actionCancel, caller: "Org1MSP", want: statusCancelled},
		{name: "payer cannot cancel", status: statusIssued, action: actionCancel, caller: "Org2MSP", wantError: true},
		{name: "cannot cancel a paid invoice", status: statusPaid, action: actionCancel, caller: "Org1MSP", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyTransition(tt.status, tt.action, tt.caller, "Org1MSP", "Org2MSP")
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected an error, got status %s", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("status = %s, want %s", got, tt.want)
			}
		})
	}
}
