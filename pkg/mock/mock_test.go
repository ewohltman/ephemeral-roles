package mock

import "testing"

func TestSession(t *testing.T) {
	session, err := Session()
	if err != nil {
		t.Fatal(err)
	}

	defer SessionClose(t, session)
}
