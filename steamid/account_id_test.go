package steamid_test

import (
	"testing"

	"github.com/13k/go-steam/steamid"
)

func TestAccountID_ID(t *testing.T) {
	subject := steamid.AccountID(0xff)
	expected := uint32(0x7f)
	actual := subject.ID()

	if actual != expected {
		t.Fatalf("expected %#x, got %#x", expected, actual)
	}
}

func TestAccountID_SetID(t *testing.T) {
	subject := steamid.NewAccountID(0xbaadf00d, 1)
	expected := steamid.NewAccountID(0xdeadbeef, 1)
	actual := subject.SetID(0xdeadbeef)

	if actual != expected {
		t.Fatalf("expected %#x, got %#x", expected, actual)
	}

	subject = steamid.NewAccountID(0xbaadf00d, 0)
	expected = steamid.NewAccountID(0xdeadbeef, 0)
	actual = subject.SetID(0xdeadbeef)

	if actual != expected {
		t.Fatalf("expected %#x, got %#x", expected, actual)
	}
}

func TestAccountID_AuthServer(t *testing.T) {
	subject := steamid.AccountID(0xff)
	expected := uint32(1)
	actual := subject.AuthServer()

	if actual != expected {
		t.Fatalf("expected %#x, got %#x", expected, actual)
	}

	subject = steamid.AccountID(0xfe)
	expected = uint32(0)
	actual = subject.AuthServer()

	if actual != expected {
		t.Fatalf("expected %#x, got %#x", expected, actual)
	}
}

func TestAccountID_SetAuthServer(t *testing.T) {
	subject := steamid.NewAccountID(0xbaadf00d, 0)
	expected := steamid.NewAccountID(0xbaadf00d, 1)
	actual := subject.SetAuthServer(1)

	if actual != expected {
		t.Fatalf("expected %#x, got %#x", expected, actual)
	}

	subject = steamid.NewAccountID(0xbaadf00d, 1)
	expected = steamid.NewAccountID(0xbaadf00d, 0)
	actual = subject.SetAuthServer(0)

	if actual != expected {
		t.Fatalf("expected %#x, got %#x", expected, actual)
	}
}
