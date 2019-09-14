package mock

import "testing"

func TestSession(t *testing.T) {
	_, err := Session()
	if err != nil {
		t.Fatal(err)
	}
}
