package media

import (
	"context"
	"testing"
)

func TestValidateRemoteMediaURLBlocksLocalAddressesByDefault(t *testing.T) {
	if err := validateRemoteMediaURL(context.Background(), "http://example.com/file.png", false); err == nil {
		t.Fatal("plain HTTP remote media should be rejected by default")
	}
	for _, target := range []string{
		"http://127.0.0.1/file.png",
		"http://localhost/file.png",
		"http://10.0.0.10/file.png",
		"http://169.254.169.254/latest/meta-data",
	} {
		if err := validateRemoteMediaURL(context.Background(), target, false); err == nil {
			t.Fatalf("validateRemoteMediaURL(%q) = nil, want blocked", target)
		}
	}
	if err := validateRemoteMediaURL(context.Background(), "http://127.0.0.1/test.png", true); err != nil {
		t.Fatalf("explicit local test source should be allowed: %v", err)
	}
}

func TestValidatedRemoteMediaTypeUsesActualBytes(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	kind, mimeType, err := validatedRemoteMediaType(MediaKindImage, "image/png", png)
	if err != nil {
		t.Fatalf("validatedRemoteMediaType(png) error = %v", err)
	}
	if kind != MediaKindImage || mimeType != "image/png" {
		t.Fatalf("png type = %q/%q", kind, mimeType)
	}

	if _, _, err := validatedRemoteMediaType(MediaKindImage, "image/png", []byte("<html>not an image</html>")); err == nil {
		t.Fatal("expected spoofed image MIME to be rejected")
	}

	if _, _, err := validatedRemoteMediaType(MediaKindImage, "video/mp4", png); err == nil {
		t.Fatal("expected declared video MIME with image bytes to be rejected")
	}

	mp4 := []byte{0, 0, 0, 24, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0, 0, 0, 0, 'i', 's', 'o', 'm'}
	kind, mimeType, err = validatedRemoteMediaType(MediaKindVideo, "video/mp4", mp4)
	if err != nil {
		t.Fatalf("validatedRemoteMediaType(mp4) error = %v", err)
	}
	if kind != MediaKindVideo || mimeType != "video/mp4" {
		t.Fatalf("mp4 type = %q/%q", kind, mimeType)
	}
}
