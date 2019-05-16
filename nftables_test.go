package nftableslib

import (
	"testing"

	"github.com/google/nftables"
)

func TestCreateTable(t *testing.T) {
	conn := InitConn()
	if conn == nil {
		t.Fatal("initialization of netlink connection failed")
	}
	conn.Tables().Create("filter", nftables.TableFamilyIPv4)
	exist := conn.Tables().Exist("filter", nftables.TableFamilyIPv4)
	if !exist {
		t.Fatalf("expected table %s of type %v to exist, but it does not", "filter", nftables.TableFamilyIPv4)
	}
}

func TestDeleteNFTable(t *testing.T) {
	conn := InitConn()
	if conn == nil {
		t.Fatal("initialization of netlink connection failed")
	}
	conn.Tables().Create("filter", nftables.TableFamilyIPv4)
	conn.Tables().Delete("filter", nftables.TableFamilyIPv4)
	exist := conn.Tables().Exist("filter", nftables.TableFamilyIPv4)
	if exist {
		t.Fatalf("expected table %s of type %v not exist, but it does", "filter", nftables.TableFamilyIPv4)
	}
}