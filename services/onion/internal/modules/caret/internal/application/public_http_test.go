package application

import (
	"net/netip"
	"testing"
)

func TestIsPublicAddrRejectsInternalNetworks(t *testing.T) {
	for _, address := range []string{"127.0.0.1", "10.0.3.4", "172.16.0.1", "192.168.1.1", "169.254.169.254", "100.64.1.1", "::1", "fe80::1", "fd00::1", "0.0.0.0", "::ffff:10.0.0.1"} {
		if isPublicAddr(netip.MustParseAddr(address)) {
			t.Errorf("%s was treated as public", address)
		}
	}
	for _, address := range []string{"1.1.1.1", "8.8.8.8", "2606:4700:4700::1111"} {
		if !isPublicAddr(netip.MustParseAddr(address)) {
			t.Errorf("%s was treated as private", address)
		}
	}
}

func TestPublicDialControlBlocksMetadataService(t *testing.T) {
	if err := publicDialControl("tcp4", "169.254.169.254:80", nil); err == nil {
		t.Fatal("the metadata service must not be reachable")
	}
	if err := publicDialControl("tcp4", "1.1.1.1:443", nil); err != nil {
		t.Fatalf("public address rejected: %v", err)
	}
}
