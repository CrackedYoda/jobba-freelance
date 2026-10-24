package dispute

import (
	"errors"
	"testing"

	"github.com/jobba/backend/internal/shared"
)

func TestAuthorizeOnlyParticipantsOrAdmin(t *testing.T) {
	d := Dispute{InitiatorID: "client", ResponderID: "worker"}
	for _, actor := range []shared.Actor{{ID: "client", Role: "client"}, {ID: "worker", Role: "worker"}, {ID: "staff", Role: "admin"}} {
		if err := authorize(d, actor); err != nil {
			t.Fatalf("authorized actor %#v rejected: %v", actor, err)
		}
	}
	if err := authorize(d, shared.Actor{ID: "stranger", Role: "worker"}); !errors.Is(err, forbidden) {
		t.Fatalf("stranger error = %v", err)
	}
}
