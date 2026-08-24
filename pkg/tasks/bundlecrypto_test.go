package tasks

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/xbapps/xbvr/pkg/models"
)

func TestBundleSecretRoundTrip(t *testing.T) {
	enc, err := EncryptBundleSecret("hunter2", `{"apiKey":"supersecret"}`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(enc, EncryptedValuePrefix) {
		t.Fatalf("missing prefix: %q", enc)
	}
	if strings.Contains(enc, "supersecret") {
		t.Fatal("ciphertext contains plaintext")
	}

	dec, err := DecryptBundleSecret("hunter2", enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != `{"apiKey":"supersecret"}` {
		t.Fatalf("round trip mismatch: %q", dec)
	}
}

func TestDecryptBundleSecretWrongPassword(t *testing.T) {
	enc, err := EncryptBundleSecret("correct", "value")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecryptBundleSecret("wrong", enc); err == nil {
		t.Fatal("expected decryption with wrong password to fail")
	}
	if _, err := DecryptBundleSecret("correct", "not-encrypted"); err == nil {
		t.Fatal("expected decryption of untagged value to fail")
	}
}

func TestEncryptBundleSecretRequiresPassword(t *testing.T) {
	if _, err := EncryptBundleSecret("", "value"); err == nil {
		t.Fatal("expected empty password to be rejected")
	}
}

func TestIsCredentialKVKey(t *testing.T) {
	cred := []string{"config", "example.com-trailers", "site-cookie", "auth-header", "example.com-scraper"}
	notCred := []string{"state", "countries", "scraper_rate_limits", "lock-scrape"}
	for _, k := range cred {
		if !IsCredentialKVKey(k) {
			t.Errorf("expected %q to be a credential key", k)
		}
	}
	for _, k := range notCred {
		if IsCredentialKVKey(k) {
			t.Errorf("expected %q to not be a credential key", k)
		}
	}
}

func TestIsEncryptedBundleValue(t *testing.T) {
	enc, err := EncryptBundleSecret("pw", "secret")
	if err != nil {
		t.Fatal(err)
	}
	if !IsEncryptedBundleValue(enc) {
		t.Fatal("expected encrypted value to be detected")
	}
	if IsEncryptedBundleValue("xbvrenc:v1:AAAA") {
		t.Fatal("never-published v1 value misdetected as encrypted")
	}
	if IsEncryptedBundleValue(`{"apiKey":"plaintext"}`) {
		t.Fatal("unencrypted value misdetected as encrypted")
	}
}

func TestBundleSecretPasswordRoundTrip(t *testing.T) {
	enc, err := EncryptBundleSecret("hunter2", "value")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(enc, EncryptedValuePrefix) {
		t.Fatalf("missing password prefix: %q", enc)
	}
	if _, err := DecryptBundleSecret("", enc); err == nil {
		t.Fatal("expected password-encrypted value to require the password")
	}
	if _, err := DecryptBundleSecret("wrong", enc); err == nil {
		t.Fatal("expected wrong password to fail")
	}
	dec, err := DecryptBundleSecret("hunter2", enc)
	if err != nil {
		t.Fatal(err)
	}
	if dec != "value" {
		t.Fatalf("round trip mismatch: %q", dec)
	}
}

func TestVerifyBundlePassword(t *testing.T) {
	enc, err := EncryptBundleSecret("hunter2", `{"apiKey":"supersecret"}`)
	if err != nil {
		t.Fatal(err)
	}
	enc2, err := EncryptBundleSecret("hunter2", "another-secret")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name     string
		kvs      []models.KV
		password string
		wantErr  bool
	}{
		{"no kvs at all", nil, "", false},
		{"unencrypted bundle needs no password", []models.KV{{Key: "config", Value: `{"apiKey":"plaintext"}`}}, "", false},
		{"mixed unencrypted values ignored", []models.KV{{Key: "state", Value: "{}"}, {Key: "config", Value: enc}}, "hunter2", false},
		{"encrypted bundle, correct password", []models.KV{{Key: "config", Value: enc}, {Key: "site-scraper", Value: enc2}}, "hunter2", false},
		{"encrypted bundle, missing password", []models.KV{{Key: "config", Value: enc}}, "", true},
		{"encrypted bundle, wrong password", []models.KV{{Key: "config", Value: enc}}, "wrong", true},
		{"second value mismatch aborts", []models.KV{{Key: "config", Value: enc}, {Key: "site-scraper", Value: enc2}}, "hunter2 ", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := VerifyBundlePassword(tc.kvs, tc.password)
			if tc.wantErr && err == nil {
				t.Fatal("expected an error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestProtectCredentialKVsRequiresPassword(t *testing.T) {
	tlog := logrus.WithField("task", "test")

	kvs := []models.KV{
		{Key: "config", Value: `{"apiKey":"supersecret"}`},
		{Key: "state", Value: `{"some":"state"}`},
	}

	// without a password the export must fail rather than drop the secrets
	if _, err := protectCredentialKVs(kvs, "", tlog); err == nil {
		t.Fatal("expected an error when exporting credentials without a password")
	}

	// with a password the credential entry is encrypted, nothing is dropped
	out, err := protectCredentialKVs(kvs, "hunter2", tlog)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	if !strings.HasPrefix(out[0].Value, EncryptedValuePrefix) {
		t.Fatalf("credential entry not encrypted: %q", out[0].Value)
	}
	if strings.Contains(out[0].Value, "supersecret") {
		t.Fatal("ciphertext contains plaintext")
	}
	if out[1].Value != `{"some":"state"}` {
		t.Fatal("non-credential entry was modified")
	}
}
