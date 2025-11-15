package util

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

var (
	subnetPrefixVerify = "_subnet-verify"
)

// VerifyDNSVerification checks if the DNS TXT record for the hostname matches the token.
// It performs a DNS lookup for <subnetPrefixVerify>.<hostname> and checks if any TXT record matches the token.
//
// Parameters:
//   - ctx: Context with timeout (the function will respect context cancellation)
//   - hostname: The hostname to verify
//   - token: The expected verification token
//   - timeout: Maximum time to wait for DNS lookup (if ctx doesn't have a timeout, this will be used)
//
// Returns:
//   - verified: true if the token matches, false otherwise
//   - message: Human-readable message describing the verification result or error
func VerifyDNSVerification(ctx context.Context, hostname, token string, timeout time.Duration) (verified bool, message string) {
	if token == "" {
		return false, "no verification token found: token should be generated automatically when hostname is first declared"
	}

	// Use context timeout if available, otherwise create one with the provided timeout
	verifyCtx := ctx
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		verifyCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	resolver := &net.Resolver{
		PreferGo: true,
	}

	// Check TXT record at <subnetPrefixVerify>.<hostname>
	verificationHostname := fmt.Sprintf("%s.%s", subnetPrefixVerify, hostname)
	txtRecords, err := resolver.LookupTXT(verifyCtx, verificationHostname)
	if err != nil {
		// Provide specific error messages based on error type
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			if dnsErr.IsNotFound {
				return false, fmt.Sprintf("DNS TXT record not found: please add a TXT record '%s.%s' with value %s in your DNS provider", subnetPrefixVerify, hostname, token)
			}
			if dnsErr.IsTimeout {
				return false, fmt.Sprintf("DNS lookup timeout: DNS server did not respond within %v, please check your DNS configuration", timeout)
			}
			if dnsErr.IsTemporary {
				return false, fmt.Sprintf("DNS temporary error: %v (this may be a transient DNS issue, please try again later)", dnsErr)
			}
			return false, fmt.Sprintf("DNS error: %v (error code: %s)", dnsErr, dnsErr.Err)
		}
		// Check for context timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return false, fmt.Sprintf("DNS lookup timeout: exceeded %v timeout, please check your DNS configuration", timeout)
		}
		return false, fmt.Sprintf("DNS lookup failed: %v (please ensure DNS is properly configured)", err)
	}

	// Check if any TXT record matches the expected token
	if len(txtRecords) == 0 {
		return false, fmt.Sprintf("no TXT records found: please add a TXT record '%s.%s' with value %s in your DNS provider", subnetPrefixVerify, hostname, token)
	}

	for _, txt := range txtRecords {
		// Remove quotes if present (some DNS servers return quoted strings)
		txt = strings.Trim(txt, "\"")
		if txt == token {
			return true, "DNS verification successful"
		}
	}

	// Token mismatch - provide helpful message
	return false, fmt.Sprintf("verification token mismatch: expected %s, but found TXT record(s): %v. Please update your DNS TXT record '%s.%s' to contain exactly: %s", token, txtRecords, subnetPrefixVerify, hostname, token)
}
