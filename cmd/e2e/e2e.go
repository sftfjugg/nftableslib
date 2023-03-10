package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"golang.org/x/sys/unix"

	"github.com/google/nftables"
	"github.com/sbezverk/nftableslib"
	"github.com/sbezverk/nftableslib/pkg/e2e/setenv"
	"github.com/sbezverk/nftableslib/pkg/e2e/validations"
)

var accept = nftableslib.ChainPolicyAccept

func init() {
	runtime.LockOSThread()
}

func main() {
	tests := []setenv.NFTablesTest{
		{
			Name:    "IPV4 ICMP Drop",
			Version: nftables.TableFamilyIPv4,
			DstNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeFilter,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookInput,
						Policy:   &accept,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_ICMP),
								Dst: &nftableslib.IPAddrSpec{
									List: []*nftableslib.IPAddr{setIPAddr("1.1.1.2")},
								},
							},
							Action: setActionVerdict(nftableslib.NFT_DROP),
						},
					},
				},
			},
			Saddr:      "1.1.1.1/24",
			Daddr:      "1.1.1.2/24",
			Validation: validations.ICMPDropTestValidation,
		},
		{
			Name:    "IPV4 Redirecting TCP port 8888 to 9999",
			Version: nftables.TableFamilyIPv4,
			DstNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: nil,
					Rules: []nftableslib.Rule{
						{
							// This rule will block ALL TCP traffic with the exception of traffic destined to port 8888
							L4: &nftableslib.L4Rule{
								L4Proto: unix.IPPROTO_TCP,
								Dst: &nftableslib.Port{
									List:  nftableslib.SetPortList([]int{8888}),
									RelOp: nftableslib.NEQ,
								},
							},
							Action: setActionVerdict(nftableslib.NFT_DROP),
						},
						{
							// Allowed TCP traffic to port 8888 will be redirected to port 9999
							L4: &nftableslib.L4Rule{
								L4Proto: unix.IPPROTO_TCP,
								Dst: &nftableslib.Port{
									List: nftableslib.SetPortList([]int{8888}),
								},
							},
							Action: setActionRedirect(9999, false),
						},
					},
				},
				{
					Name: "chain-2",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeNAT,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookPrerouting,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_TCP),
							},
							Action: setActionVerdict(unix.NFT_JUMP, "chain-1"),
						},
					},
				},
			},
			Saddr:      "1.1.1.1/24",
			Daddr:      "1.1.1.2/24",
			Validation: validations.TCPPortRedirectValidation,
		},
		{
			Name:    "IPV6 Redirecting TCP port 8888 to 9999",
			Version: nftables.TableFamilyIPv6,
			DstNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: nil,
					Rules: []nftableslib.Rule{
						{
							// This rule will block ALL TCP traffic with the exception of traffic destined to port 8888
							L4: &nftableslib.L4Rule{
								L4Proto: unix.IPPROTO_TCP,
								Dst: &nftableslib.Port{
									List:  nftableslib.SetPortList([]int{8888}),
									RelOp: nftableslib.NEQ,
								},
							},
							Action: setActionVerdict(nftableslib.NFT_DROP),
						},
						{
							// Allowed TCP traffic to port 8888 will be redirected to port 9999
							L4: &nftableslib.L4Rule{
								L4Proto: unix.IPPROTO_TCP,
								Dst: &nftableslib.Port{
									List: nftableslib.SetPortList([]int{8888}),
								},
							},
							Action: setActionRedirect(9999, false),
						},
					},
				},
				{
					Name: "chain-2",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeNAT,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookPrerouting,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_TCP),
							},
							Action: setActionVerdict(unix.NFT_JUMP, "chain-1"),
						},
					},
				},
			},
			Saddr:      "2001:1::1/64",
			Daddr:      "2001:1::2/64",
			Validation: validations.TCPPortRedirectValidation,
		},
		{
			Name:    "IPV4 TCP SNAT",
			Version: nftables.TableFamilyIPv4,
			SrcNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeNAT,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookPostrouting,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_TCP),
							},
							Action: setSNAT(&nftableslib.NATAttributes{
								L3Addr: [2]*nftableslib.IPAddr{setIPAddr("5.5.5.5")},
								Port:   [2]uint16{7777},
							}),
						},
					},
				},
			},
			Saddr:        "1.1.1.1/24",
			Daddr:        "1.1.1.2/24",
			Validation:   validations.IPv4TCPSNATValidation,
			DebugNFRules: false,
		},
		{
			Name:    "IPV6 TCP SNAT",
			Version: nftables.TableFamilyIPv6,
			SrcNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeNAT,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookPostrouting,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_TCP),
							},
							Action: setSNAT(&nftableslib.NATAttributes{
								L3Addr: [2]*nftableslib.IPAddr{setIPAddr("2001:1234::1")},
								Port:   [2]uint16{7777},
							})},
					},
				},
			},
			Saddr:      "2001:1::1/64",
			Daddr:      "2001:1::2/64",
			Validation: validations.IPv6TCPSNATValidation,
		},
		{
			Name:    "IPV4 UDP SNAT",
			Version: nftables.TableFamilyIPv4,
			SrcNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeNAT,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookPostrouting,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_UDP),
							},
							Action: setSNAT(&nftableslib.NATAttributes{
								L3Addr: [2]*nftableslib.IPAddr{setIPAddr("5.5.5.5")},
								Port:   [2]uint16{7777},
							}),
						},
					},
				},
			},
			Saddr:        "1.1.1.1/24",
			Daddr:        "1.1.1.2/24",
			Validation:   validations.IPv4UDPSNATValidation,
			DebugNFRules: false,
		},
		{
			Name:    "IPV6 UDP SNAT",
			Version: nftables.TableFamilyIPv6,
			SrcNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeNAT,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookPostrouting,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_UDP),
							},
							Action: setSNAT(&nftableslib.NATAttributes{
								L3Addr: [2]*nftableslib.IPAddr{setIPAddr("2001:1234::1")},
								Port:   [2]uint16{7777},
							})},
					},
				},
			},
			Saddr:      "2001:1::1/64",
			Daddr:      "2001:1::2/64",
			Validation: validations.IPv6UDPSNATValidation,
		},
		{
			Name:    "IPV6 ICMP Drop",
			Version: nftables.TableFamilyIPv6,
			DstNFRules: []setenv.TestChain{
				{
					Name: "chain-1",
					Attr: &nftableslib.ChainAttributes{
						Type:     nftables.ChainTypeFilter,
						Priority: nftables.ChainPriorityFirst,
						Hook:     nftables.ChainHookInput,
						Policy:   &accept,
					},
					Rules: []nftableslib.Rule{
						{
							L3: &nftableslib.L3Rule{
								Protocol: nftableslib.L3Protocol(unix.IPPROTO_ICMPV6),
								Dst: &nftableslib.IPAddrSpec{
									List: []*nftableslib.IPAddr{setIPAddr("2001:1::2")},
								},
							},
							Action: setActionVerdict(nftableslib.NFT_DROP),
						},
					},
				},
			},
			Saddr:      "2001:1::1/64",
			Daddr:      "2001:1::2/64",
			Validation: validations.ICMPDropTestValidation,
		},
	}

	memProf, err := os.Create("/tmp/heap.out")
	if err != nil {
		os.Exit(1)
	}
	defer memProf.Close()
	//	if err := pprof.WriteHeapProfile(memProf); err != nil {
	//		fmt.Printf("Error writing memory profile with error: %+v\n", err)
	//	}
	for _, tt := range tests {

		fmt.Printf("+++ Starting test: \"%s\" \n", tt.Name)
		t, err := setenv.NewP2PTestEnv(tt.Version, tt.Saddr, tt.Daddr)
		if err != nil {
			fmt.Printf("--- Test: \"%s\" failed with error: %+v\n", tt.Name, err)
			os.Exit(1)
		}
		defer t.Cleanup()
		// Get allocated namesapces and prepared ip addresses
		ns := t.GetNamespace()
		ip := t.GetIPs()

		// Initial connectivity test before applying any nftables rules
		if err := setenv.TestICMP(ns[0], tt.Version, ip[0], ip[1]); err != nil {
			fmt.Printf("--- Test: \"%s\" failed during initial connectivity test with error: %+v\n", tt.Name, err)
			os.Exit(1)
		}
		if tt.SrcNFRules != nil {
			if _, err := setenv.NFTablesSet(setenv.MakeTablesInterface(ns[0]), tt.Version, tt.SrcNFRules, tt.DebugNFRules); err != nil {
				fmt.Printf("--- Test: \"%s\" failed to setup nftables table/chain/rule in a source namespace with error: %+v\n", tt.Name, err)
				os.Exit(1)
			}
		}
		if tt.DstNFRules != nil {
			if _, err := setenv.NFTablesSet(setenv.MakeTablesInterface(ns[1]), tt.Version, tt.DstNFRules, tt.DebugNFRules); err != nil {
				fmt.Printf("--- Test: \"%s\" failed to setup nftables table/chain/rule in a destination namespace with error: %+v\n", tt.Name, err)
				os.Exit(1)
			}
		}
		// Check if test's validation is set and execute validation.
		if tt.Validation != nil {
			if err := tt.Validation(tt.Version, ns, ip); err != nil {
				fmt.Printf("--- Test: \"%s\" failed validation error: %+v\n", tt.Name, err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("--- Test: \"%s\" has no validation, test without validation is not allowed\n", tt.Name)
			os.Exit(1)
		}
		fmt.Printf("+++ Finished test: \"%s\" successfully.\n", tt.Name)
	}
	fmt.Printf("+++ Starting test: Sync() \n")
	// Testing Sync feature, in a namespace a set of rules will be created and programmed, then tables/chains/rules in
	// memory removed, Sync is supposed to learn and rebuild in-memory data structures based on discovered in the namesapce
	// nftables information.
	if err := testSync(); err != nil {
		fmt.Printf("--- Test: Sync failed with error: %+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("+++ Finished test: Sync() successfully.\n")

	if err := pprof.WriteHeapProfile(memProf); err != nil {
		fmt.Printf("Error writing memory profile with error: %+v\n", err)
	}
}

func setActionVerdict(key int, chain ...string) *nftableslib.RuleAction {
	ra, err := nftableslib.SetVerdict(key, chain...)
	if err != nil {
		fmt.Printf("failed to SetVerdict with error: %+v\n", err)
		return nil
	}
	return ra
}

func setActionRedirect(port int, tproxy bool) *nftableslib.RuleAction {
	ra, err := nftableslib.SetRedirect(port, tproxy)
	if err != nil {
		fmt.Printf("failed to SetRedirect with error: %+v", err)
		return nil
	}
	return ra
}

func setIPAddr(addr string) *nftableslib.IPAddr {
	a, err := nftableslib.NewIPAddr(addr)
	if err != nil {
		fmt.Printf("error %+v return from NewIPAddr for address: %s\n", err, addr)
		return nil
	}
	return a
}

func setSNAT(attrs *nftableslib.NATAttributes) *nftableslib.RuleAction {
	ra, err := nftableslib.SetSNAT(attrs)
	if err != nil {
		fmt.Printf("error %+v return from SetSNAT call\n", err)
		return nil
	}
	return ra
}

func setDNAT(attrs *nftableslib.NATAttributes) *nftableslib.RuleAction {
	ra, err := nftableslib.SetDNAT(attrs)
	if err != nil {
		fmt.Printf("error %+v return from SetSNAT call\n", err)
		return nil
	}
	return ra
}
