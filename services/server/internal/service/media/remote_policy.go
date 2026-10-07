package media

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
)

var blockedRemoteMediaPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("2001:db8::/32"),
}

func validateRemoteMediaURL(ctx context.Context, rawURL string, allowUnsafeLocal bool) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return fmt.Errorf("invalid remote asset url: %w", err)
	}
	if allowUnsafeLocal {
		if parsed.Scheme != "https" && parsed.Scheme != "http" {
			return fmt.Errorf("remote asset url scheme must be http or https")
		}
	} else if parsed.Scheme != "https" {
		return fmt.Errorf("remote asset url must use https")
	}
	if parsed.User != nil {
		return fmt.Errorf("remote asset url must not include user info")
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("remote asset url host is empty")
	}
	if allowUnsafeLocal {
		return nil
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return fmt.Errorf("remote asset url resolves to a local host")
	}
	if ip := net.ParseIP(host); ip != nil {
		return validateRemoteMediaIP(ip)
	}
	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("resolving remote asset host: %w", err)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("remote asset host has no resolved address")
	}
	for _, address := range addresses {
		if err := validateRemoteMediaIP(address.IP); err != nil {
			return fmt.Errorf("remote asset host %s: %w", host, err)
		}
	}
	return nil
}

func validateRemoteMediaIP(ip net.IP) error {
	if ip == nil {
		return fmt.Errorf("invalid remote address")
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() {
		return fmt.Errorf("remote address %s is not allowed", ip.String())
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return fmt.Errorf("invalid remote address %s", ip.String())
	}
	addr = addr.Unmap()
	for _, prefix := range blockedRemoteMediaPrefixes {
		if prefix.Contains(addr) {
			return fmt.Errorf("remote address %s is reserved", addr.String())
		}
	}
	return nil
}

func remoteMediaHTTPClient(allowUnsafeLocal bool) *http.Client {
	client := *mediaAssetHTTPClient
	client.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("remote asset redirect limit exceeded")
		}
		return validateRemoteMediaURL(request.Context(), request.URL.String(), allowUnsafeLocal)
	}
	return &client
}

func validatedRemoteMediaType(kind string, declared string, data []byte) (string, string, error) {
	if len(data) == 0 {
		return "", "", fmt.Errorf("remote media payload is empty")
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	declared = normalizeRemoteMediaMIME(declared)
	detected := detectRemoteMediaMIME(data)
	detectedKind := sharedKindFromRemoteMIME(detected)
	if detectedKind == "" {
		return "", "", fmt.Errorf("remote media signature is unsupported")
	}
	if kind != "" && kind != detectedKind {
		return "", "", fmt.Errorf("remote media kind mismatch: expected %s, detected %s", kind, detectedKind)
	}
	if kind == "" {
		kind = detectedKind
	}
	if declaredKind := sharedKindFromRemoteMIME(declared); declaredKind != "" && declaredKind != detectedKind {
		return "", "", fmt.Errorf("remote media MIME mismatch: declared %s, detected %s", declared, detected)
	}
	return kind, detected, nil
}

func normalizeRemoteMediaMIME(value string) string {
	return strings.ToLower(strings.TrimSpace(strings.Split(value, ";")[0]))
}

func sharedKindFromRemoteMIME(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return MediaKindImage
	case strings.HasPrefix(mimeType, "video/"):
		return MediaKindVideo
	case strings.HasPrefix(mimeType, "audio/"):
		return MediaKindAudio
	default:
		return ""
	}
}

func detectRemoteMediaMIME(data []byte) string {
	sniffed := normalizeRemoteMediaMIME(http.DetectContentType(data))
	if sharedKindFromRemoteMIME(sniffed) != "" {
		return sniffed
	}
	if len(data) >= 12 && string(data[4:8]) == "ftyp" {
		brand := string(data[8:12])
		switch brand {
		case "M4A ", "M4B ", "M4P ":
			return "audio/mp4"
		default:
			return "video/mp4"
		}
	}
	if len(data) >= 4 && data[0] == 0x1a && data[1] == 0x45 && data[2] == 0xdf && data[3] == 0xa3 {
		return "video/webm"
	}
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE" {
		return "audio/wav"
	}
	if len(data) >= 4 && string(data[0:4]) == "OggS" {
		return "audio/ogg"
	}
	if len(data) >= 3 && string(data[0:3]) == "ID3" {
		return "audio/mpeg"
	}
	if len(data) >= 2 && data[0] == 0xff && data[1]&0xe0 == 0xe0 {
		return "audio/mpeg"
	}
	if len(data) >= 4 && binary.BigEndian.Uint32(data[:4]) == 0x89504e47 {
		return "image/png"
	}
	return ""
}
