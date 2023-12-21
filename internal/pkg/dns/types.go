package dns

import (
	"fmt"
	"strings"
)

// Type represents a DNS record type.
type Type uint16

const (
	TypeA          Type = 1
	TypeNS         Type = 2
	TypeMD         Type = 3
	TypeMF         Type = 4
	TypeCNAME      Type = 5
	TypeSOA        Type = 6
	TypeMB         Type = 7
	TypeMG         Type = 8
	TypeMR         Type = 9
	TypeNULL       Type = 10
	TypePTR        Type = 12
	TypeHINFO      Type = 13
	TypeMINFO      Type = 14
	TypeMX         Type = 15
	TypeTXT        Type = 16
	TypeRP         Type = 17
	TypeAFSDB      Type = 18
	TypeX25        Type = 19
	TypeISDN       Type = 20
	TypeRT         Type = 21
	TypeNSAPPTR    Type = 23
	TypeSIG        Type = 24
	TypeKEY        Type = 25
	TypePX         Type = 26
	TypeGPOS       Type = 27
	TypeAAAA       Type = 28
	TypeLOC        Type = 29
	TypeNXT        Type = 30
	TypeEID        Type = 31
	TypeNIMLOC     Type = 32
	TypeSRV        Type = 33
	TypeATMA       Type = 34
	TypeNAPTR      Type = 35
	TypeKX         Type = 36
	TypeCERT       Type = 37
	TypeDNAME      Type = 39
	TypeOPT        Type = 41
	TypeAPL        Type = 42
	TypeDS         Type = 43
	TypeSSHFP      Type = 44
	TypeIPSECKEY   Type = 45
	TypeRRSIG      Type = 46
	TypeNSEC       Type = 47
	TypeDNSKEY     Type = 48
	TypeDHCID      Type = 49
	TypeNSEC3      Type = 50
	TypeNSEC3PARAM Type = 51
	TypeTLSA       Type = 52
	TypeSMIMEA     Type = 53
	TypeHIP        Type = 55
	TypeRKEY       Type = 57
	TypeTALINK     Type = 58
	TypeCDS        Type = 59
	TypeCDNSKEY    Type = 60
	TypeOPENPGPKEY Type = 61
	TypeCSYNC      Type = 62
	TypeZONEMD     Type = 63
	TypeSVCB       Type = 64
	TypeHTTPS      Type = 65
	TypeSPF        Type = 99
	TypeNID        Type = 104
	TypeL32        Type = 105
	TypeL64        Type = 106
	TypeLP         Type = 107
	TypeEUI48      Type = 108
	TypeEUI64      Type = 109
	TypeURI        Type = 256
	TypeCAA        Type = 257
	TypeAVC        Type = 258
	TypeAMTRELAY   Type = 260

	TypeTKEY Type = 249
	TypeTSIG Type = 250

	// valid Question.Qtype only.
	TypeIXFR  Type = 251
	TypeAXFR  Type = 252
	TypeMAILB Type = 253
	TypeMAILA Type = 254
	TypeANY   Type = 255

	TypeTA  Type = 32768
	TypeDLV Type = 32769
)

var recordTypeNames = map[Type]string{
	TypeA:          "A",
	TypeAAAA:       "AAAA",
	TypeAFSDB:      "AFSDB",
	TypeAMTRELAY:   "AMTRELAY",
	TypeANY:        "ANY",
	TypeAPL:        "APL",
	TypeATMA:       "ATMA",
	TypeAVC:        "AVC",
	TypeAXFR:       "AXFR",
	TypeCAA:        "CAA",
	TypeCDNSKEY:    "CDNSKEY",
	TypeCDS:        "CDS",
	TypeCERT:       "CERT",
	TypeCNAME:      "CNAME",
	TypeCSYNC:      "CSYNC",
	TypeDHCID:      "DHCID",
	TypeDLV:        "DLV",
	TypeDNAME:      "DNAME",
	TypeDNSKEY:     "DNSKEY",
	TypeDS:         "DS",
	TypeEID:        "EID",
	TypeEUI48:      "EUI48",
	TypeEUI64:      "EUI64",
	TypeGPOS:       "GPOS",
	TypeHINFO:      "HINFO",
	TypeHIP:        "HIP",
	TypeHTTPS:      "HTTPS",
	TypeIPSECKEY:   "IPSECKEY",
	TypeISDN:       "ISDN",
	TypeIXFR:       "IXFR",
	TypeKEY:        "KEY",
	TypeKX:         "KX",
	TypeL32:        "L32",
	TypeL64:        "L64",
	TypeLOC:        "LOC",
	TypeLP:         "LP",
	TypeMAILA:      "MAILA",
	TypeMAILB:      "MAILB",
	TypeMB:         "MB",
	TypeMD:         "MD",
	TypeMF:         "MF",
	TypeMG:         "MG",
	TypeMINFO:      "MINFO",
	TypeMR:         "MR",
	TypeMX:         "MX",
	TypeNAPTR:      "NAPTR",
	TypeNID:        "NID",
	TypeNIMLOC:     "NIMLOC",
	TypeNS:         "NS",
	TypeNSAPPTR:    "NSAP-PTR",
	TypeNSEC3:      "NSEC3",
	TypeNSEC3PARAM: "NSEC3PARAM",
	TypeNSEC:       "NSEC",
	TypeNULL:       "NULL",
	TypeNXT:        "NXT",
	TypeOPENPGPKEY: "OPENPGPKEY",
	TypeOPT:        "OPT",
	TypePTR:        "PTR",
	TypePX:         "PX",
	TypeRKEY:       "RKEY",
	TypeRP:         "RP",
	TypeRRSIG:      "RRSIG",
	TypeRT:         "RT",
	TypeSIG:        "SIG",
	TypeSMIMEA:     "SMIMEA",
	TypeSOA:        "SOA",
	TypeSPF:        "SPF",
	TypeSRV:        "SRV",
	TypeSSHFP:      "SSHFP",
	TypeSVCB:       "SVCB",
	TypeTA:         "TA",
	TypeTALINK:     "TALINK",
	TypeTKEY:       "TKEY",
	TypeTLSA:       "TLSA",
	TypeTSIG:       "TSIG",
	TypeTXT:        "TXT",
	TypeURI:        "URI",
	TypeX25:        "X25",
	TypeZONEMD:     "ZONEMD",
}

var recordTypeDescriptions = map[Type]string{
	TypeA:          "Host address",
	TypeAAAA:       "IP6 Address",
	TypeAFSDB:      "AFS Data Base location",
	TypeAMTRELAY:   "Automatic Multicast Tunneling Relay",
	TypeANY:        "All DNS records available",
	TypeAPL:        "APL",
	TypeATMA:       "ATM Address",
	TypeAVC:        "Application Visibility and Control",
	TypeAXFR:       "Transfer of an entire zone",
	TypeCAA:        "Certification Authority Restriction",
	TypeCDNSKEY:    "DNSKEY(s) the Child wants reflected in DS",
	TypeCDS:        "Child DS",
	TypeCERT:       "CERT",
	TypeCNAME:      "Canonical name for an alias",
	TypeCSYNC:      "Child-To-Parent Synchronization",
	TypeDHCID:      "DHCID",
	TypeDLV:        "DNSSEC Lookaside Validation (OBSOLETE)",
	TypeDNAME:      "DNAME",
	TypeDNSKEY:     "DNSKEY",
	TypeDS:         "Delegation Signer",
	TypeEID:        "Endpoint Identifier",
	TypeEUI48:      "EUI-48 address",
	TypeEUI64:      "EUI-64 address",
	TypeGPOS:       "Geographical Position",
	TypeHINFO:      "Host information",
	TypeHIP:        "Host Identity Protocol",
	TypeHTTPS:      "SVCB-compatible type for use with HTTP",
	TypeIPSECKEY:   "IPSECKEY",
	TypeISDN:       "ISDN address",
	TypeIXFR:       "Incremental transfer",
	TypeKEY:        "Security key",
	TypeKX:         "Key Exchanger",
	TypeL32:        "ILNP 32-bits Locator",
	TypeL64:        "ILNP 64-bits Locator",
	TypeLOC:        "Location Information",
	TypeLP:         "ILNP subnetwork name",
	TypeMAILA:      "Mail agent RRs (OBSOLETE - see MX)",
	TypeMAILB:      "Mailbox-related RRs (MB, MG or MR)",
	TypeMB:         "Mailbox domain name (EXPERIMENTAL)",
	TypeMD:         "Mail destination (OBSOLETE - use MX)",
	TypeMF:         "Mail forwarder (OBSOLETE - use MX)",
	TypeMG:         "Mail group member (EXPERIMENTAL)",
	TypeMINFO:      "Mailbox or mail list information",
	TypeMR:         "Mail rename domain name (EXPERIMENTAL)",
	TypeMX:         "Mail exchange",
	TypeNAPTR:      "Naming Authority Pointer",
	TypeNID:        "ILNP node identifier",
	TypeNIMLOC:     "Nimrod Locator",
	TypeNS:         "Authoritative name server",
	TypeNSAPPTR:    "Domain name pointer, NSAP style (DEPRECATED)",
	TypeNSEC3:      "NSEC3",
	TypeNSEC3PARAM: "NSEC3PARAM",
	TypeNSEC:       "NSEC",
	TypeNULL:       "Null RR (EXPERIMENTAL)",
	TypeNXT:        "Next Domain (OBSOLETE)",
	TypeOPENPGPKEY: "OpenPGP Key",
	TypeOPT:        "OPT",
	TypePTR:        "Domain name pointer",
	TypePX:         "X.400 mail mapping information",
	TypeRKEY:       "RKEY",
	TypeRP:         "Responsible Person",
	TypeRRSIG:      "RRSIG",
	TypeRT:         "Route Through",
	TypeSIG:        "Security signature",
	TypeSMIMEA:     "S/MIME cert association",
	TypeSOA:        "Marks the start of a zone of authority",
	TypeSPF:        "SPF records (OBSOLETE, in TXT)",
	TypeSRV:        "Server Selection",
	TypeSSHFP:      "SSH Key Fingerprint",
	TypeSVCB:       "General-purpose service binding",
	TypeTA:         "DNSSEC Trust Authorities",
	TypeTALINK:     "Trust Anchor LINK",
	TypeTKEY:       "Transaction Key",
	TypeTLSA:       "TLSA",
	TypeTSIG:       "Transaction Signature",
	TypeTXT:        "Text strings",
	TypeURI:        "URI",
	TypeX25:        "X.25 PSDN address",
	TypeZONEMD:     "Message Digest Over Zone Data",
}

var recordTypeURLs = map[Type][]string{
	TypeA:          {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeAAAA:       {"https://datatracker.ietf.org/doc/html/rfc3596"},
	TypeAFSDB:      {"https://datatracker.ietf.org/doc/html/rfc1183", "https://datatracker.ietf.org/doc/html/rfc5864"},
	TypeAMTRELAY:   {"https://datatracker.ietf.org/doc/html/rfc8777"},
	TypeANY:        {"https://datatracker.ietf.org/doc/html/rfc8482"},
	TypeAPL:        {"https://datatracker.ietf.org/doc/html/rfc3123"},
	TypeATMA:       {"https://datatracker.ietf.org/doc/html/draft-lewis-dnsnxt-semantics-01"},
	TypeAVC:        {"https://www.cisco.com/c/en/us/td/docs/switches/lan/catalyst3650/software/release/16-3/configuration_guide/b_163_consolidated_3650_cg/b_163_consolidated_3650_cg_chapter_01111010.pdf"},
	TypeAXFR:       {"https://datatracker.ietf.org/doc/html/rfc1035", "https://datatracker.ietf.org/doc/html/rfc5936"},
	TypeCAA:        {"https://datatracker.ietf.org/doc/html/rfc8659"},
	TypeCDNSKEY:    {"https://datatracker.ietf.org/doc/html/rfc7344"},
	TypeCDS:        {"https://datatracker.ietf.org/doc/html/rfc7344"},
	TypeCERT:       {"https://datatracker.ietf.org/doc/html/rfc4398"},
	TypeCNAME:      {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeCSYNC:      {"https://datatracker.ietf.org/doc/html/rfc7477"},
	TypeDHCID:      {"https://datatracker.ietf.org/doc/html/rfc4701"},
	TypeDLV:        {"https://datatracker.ietf.org/doc/html/rfc8749", "https://datatracker.ietf.org/doc/html/rfc4431"},
	TypeDNAME:      {"https://datatracker.ietf.org/doc/html/rfc6672"},
	TypeDNSKEY:     {"https://datatracker.ietf.org/doc/html/rfc4034"},
	TypeDS:         {"https://datatracker.ietf.org/doc/html/rfc4034"},
	TypeEID:        {"http://ana-3.lcs.mit.edu/~jnc/nimrod/dns.txt"},
	TypeEUI48:      {"https://datatracker.ietf.org/doc/html/rfc7043"},
	TypeEUI64:      {"https://datatracker.ietf.org/doc/html/rfc7043"},
	TypeGPOS:       {"https://datatracker.ietf.org/doc/html/rfc1712"},
	TypeHINFO:      {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeHIP:        {"https://datatracker.ietf.org/doc/html/rfc8005"},
	TypeHTTPS:      {"https://datatracker.ietf.org/doc/html/rfc9460"},
	TypeIPSECKEY:   {"https://datatracker.ietf.org/doc/html/rfc4025"},
	TypeISDN:       {"https://datatracker.ietf.org/doc/html/rfc1183"},
	TypeIXFR:       {"https://datatracker.ietf.org/doc/html/rfc1995"},
	TypeKEY:        {"https://datatracker.ietf.org/doc/html/rfc2536", "https://datatracker.ietf.org/doc/html/rfc2539", "https://datatracker.ietf.org/doc/html/rfc3110", "https://datatracker.ietf.org/doc/html/rfc4034"},
	TypeKX:         {"https://datatracker.ietf.org/doc/html/rfc2230"},
	TypeL32:        {"https://datatracker.ietf.org/doc/html/rfc6742"},
	TypeL64:        {"https://datatracker.ietf.org/doc/html/rfc6742"},
	TypeLOC:        {"https://datatracker.ietf.org/doc/html/rfc1876"},
	TypeLP:         {"https://datatracker.ietf.org/doc/html/rfc6742"},
	TypeMAILA:      {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMAILB:      {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMB:         {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMD:         {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMF:         {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMG:         {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMINFO:      {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMR:         {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeMX:         {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeNAPTR:      {"https://datatracker.ietf.org/doc/html/rfc3403"},
	TypeNID:        {"https://datatracker.ietf.org/doc/html/rfc6742"},
	TypeNIMLOC:     {"http://ana-3.lcs.mit.edu/~jnc/nimrod/dns.txt"},
	TypeNS:         {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeNSAPPTR:    {"https://datatracker.ietf.org/doc/html/rfc1706"},
	TypeNSEC3:      {"https://datatracker.ietf.org/doc/html/rfc5155", "https://datatracker.ietf.org/doc/html/rfc9077"},
	TypeNSEC3PARAM: {"https://datatracker.ietf.org/doc/html/rfc5155"},
	TypeNSEC:       {"https://datatracker.ietf.org/doc/html/rfc4034", "https://datatracker.ietf.org/doc/html/rfc9077"},
	TypeNULL:       {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeNXT:        {"https://datatracker.ietf.org/doc/html/rfc2535", "https://datatracker.ietf.org/doc/html/rfc3755"},
	TypeOPENPGPKEY: {"https://datatracker.ietf.org/doc/html/rfc7929"},
	TypeOPT:        {"https://datatracker.ietf.org/doc/html/rfc3225", "https://datatracker.ietf.org/doc/html/rfc6891"},
	TypePTR:        {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypePX:         {"https://datatracker.ietf.org/doc/html/rfc2163"},
	TypeRKEY:       {"https://datatracker.ietf.org/doc/html/draft-reid-dnsext-rkey-00"},
	TypeRP:         {"https://datatracker.ietf.org/doc/html/rfc1183"},
	TypeRRSIG:      {"https://datatracker.ietf.org/doc/html/rfc4034"},
	TypeRT:         {"https://datatracker.ietf.org/doc/html/rfc1183"},
	TypeSIG:        {"https://datatracker.ietf.org/doc/html/rfc2536", "https://datatracker.ietf.org/doc/html/rfc2931", "https://datatracker.ietf.org/doc/html/rfc3110", "https://datatracker.ietf.org/doc/html/rfc4034"},
	TypeSMIMEA:     {"https://datatracker.ietf.org/doc/html/rfc8162"},
	TypeSOA:        {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeSPF:        {"https://datatracker.ietf.org/doc/html/rfc7208"},
	TypeSRV:        {"https://datatracker.ietf.org/doc/html/rfc2782"},
	TypeSSHFP:      {"https://datatracker.ietf.org/doc/html/rfc4255"},
	TypeSVCB:       {"https://datatracker.ietf.org/doc/html/rfc9460"},
	TypeTA:         {"http://cameo.library.cmu.edu/"},
	TypeTALINK:     {"https://datatracker.ietf.org/doc/html/draft-ietf-dnsop-dnssec-trust-history-00"},
	TypeTKEY:       {"https://datatracker.ietf.org/doc/html/rfc2930"},
	TypeTLSA:       {"https://datatracker.ietf.org/doc/html/rfc6698"},
	TypeTSIG:       {"https://datatracker.ietf.org/doc/html/rfc8945"},
	TypeTXT:        {"https://datatracker.ietf.org/doc/html/rfc1035"},
	TypeURI:        {"https://datatracker.ietf.org/doc/html/rfc7553"},
	TypeX25:        {"https://datatracker.ietf.org/doc/html/rfc1183"},
	TypeZONEMD:     {"https://datatracker.ietf.org/doc/html/rfc8976"},
}

// GetType returns the Type corresponding to the given string.
func GetType(v string) (Type, error) {
	vv := strings.ToUpper(v)
	for k, v := range recordTypeNames {
		if vv == v {
			return k, nil
		}
	}
	return 0, fmt.Errorf("unknown or unhandled record type %s", vv)
}

// GetTypeNames returns the list of all record type names.
func GetTypeNames() []string {
	names := make([]string, len(recordTypeNames))
	i := 0
	for _, v := range recordTypeNames {
		names[i] = v
		i++
	}
	return names
}

// GetTypeName returns the name for given type.
func GetTypeName(t Type) string {
	v, ok := recordTypeNames[t]
	if ok && len(v) > 0 {
		return v
	}
	return "N/A"
}

// GetTypeDescription returns the description of a record type or "N/A" if not described.
func GetTypeDescription(t Type) string {
	v, ok := recordTypeDescriptions[t]
	if ok && len(v) > 0 {
		return v
	}
	return "N/A"
}

// GetTypeURLs returns the url for a record type.
func GetTypeURLs(t Type) []string {
	v, ok := recordTypeURLs[t]
	if ok && len(v) > 0 {
		return v
	}
	return nil
}
