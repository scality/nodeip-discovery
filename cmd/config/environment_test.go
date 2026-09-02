package config

import (
	"net/netip"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/scality/go-errors"
)

var _ = Describe("validateCIDRs", func() {
	DescribeTable("valid CIDR lists",
		func(cidrList string, expected []netip.Prefix) {
			got, err := validateCIDRs(cidrList)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(expected))
		},
		Entry("valid CIDR", "192.168.1.0/24",
			[]netip.Prefix{netip.MustParsePrefix("192.168.1.0/24")}),
		Entry("masks host bits", "192.168.1.5/24",
			[]netip.Prefix{netip.MustParsePrefix("192.168.1.0/24")}),
		Entry("leading-zero prefix length", "192.168.1.0/024",
			[]netip.Prefix{netip.MustParsePrefix("192.168.1.0/24")}),
		Entry("leading-zero prefix length is honoured, not just parsed", "192.168.1.0/023",
			[]netip.Prefix{netip.MustParsePrefix("192.168.0.0/23")}),
		Entry("multiple CIDRs, trimming whitespace", " 10.0.0.0/8 , 192.168.0.0/16 ",
			[]netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("192.168.0.0/16"),
			}),
		Entry("skips empty list entries", "10.0.0.0/8,,192.168.0.0/16",
			[]netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("192.168.0.0/16"),
			}),
	)

	DescribeTable("invalid CIDR lists",
		func(cidrList string) {
			_, err := validateCIDRs(cidrList)

			Expect(errors.Is(err, ErrConfigurationLoading)).To(BeTrue())
		},
		Entry("prefix length out of range", "192.168.1.0/33"),
		Entry("IPv6 CIDR", "fd00::/64"),
		Entry("IPv4-mapped IPv6 CIDR", "::ffff:192.0.2.0/120"),
		Entry("IPv4-mapped IPv6 CIDR with IPv4 prefix length", "::ffff:192.0.2.0/24"),
		Entry("address without a prefix length", "10.0.0.1"),
		Entry("trailing slash without a prefix length", "10.0.0.0/"),
		Entry("not a CIDR", "not-a-cidr"),
		Entry("empty", ""),
		Entry("whitespace only", "   "),
		Entry("separators only", ",,"),
	)
})

var _ = Describe("validateExcludeIPs", func() {
	DescribeTable("valid exclude lists",
		func(excludeIPs string, expected []netip.Addr) {
			got, err := validateExcludeIPs(excludeIPs)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(expected))
		},
		Entry("single IP", "10.0.0.1",
			[]netip.Addr{netip.MustParseAddr("10.0.0.1")}),
		Entry("multiple IPs, trimming whitespace", " 10.0.0.1 , 10.0.0.2 ",
			[]netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
				netip.MustParseAddr("10.0.0.2"),
			}),
		Entry("skips empty list entries", "10.0.0.1,,10.0.0.2",
			[]netip.Addr{
				netip.MustParseAddr("10.0.0.1"),
				netip.MustParseAddr("10.0.0.2"),
			}),
	)

	// Unlike the CIDR list, an absent exclude list is valid: -exclude-ips is optional.
	DescribeTable("empty exclude lists are accepted",
		func(excludeIPs string) {
			got, err := validateExcludeIPs(excludeIPs)

			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeEmpty())
		},
		Entry("empty", ""),
		Entry("whitespace only", "   "),
		Entry("separators only", ",,"),
	)

	DescribeTable("invalid exclude lists",
		func(excludeIPs string) {
			_, err := validateExcludeIPs(excludeIPs)

			Expect(errors.Is(err, ErrConfigurationLoading)).To(BeTrue())
		},
		Entry("IPv6 address", "fd00::1"),
		Entry("IPv4-mapped IPv6 address", "::ffff:192.0.2.1"),
		Entry("IPv6 address with a zone", "fe80::1%eth0"),
		Entry("leading-zero octet", "192.168.001.1"),
		Entry("CIDR instead of an address", "10.0.0.0/8"),
		Entry("not an IP", "not-an-ip"),
	)
})
