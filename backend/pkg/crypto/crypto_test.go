package crypto

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := NewKey("unit-test-secret")
	cipher, err := Encrypt(key, "Sup3rS3cret#2026")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if cipher == "Sup3rS3cret#2026" {
		t.Fatal("ciphertext must not equal plaintext")
	}
	plain, err := Decrypt(key, cipher)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if plain != "Sup3rS3cret#2026" {
		t.Fatalf("round-trip mismatch: got %q", plain)
	}
}

func TestEncryptRandomNonce(t *testing.T) {
	key := NewKey("unit-test-secret")
	a, _ := Encrypt(key, "same-plaintext")
	b, _ := Encrypt(key, "same-plaintext")
	if a == b {
		t.Fatal("two encryptions of same plaintext must differ (random nonce)")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	cipher, _ := Encrypt(NewKey("secret-a"), "hello")
	if _, err := Decrypt(NewKey("secret-b"), cipher); err == nil {
		t.Fatal("expected error with wrong key")
	}
}

func TestDecryptTampered(t *testing.T) {
	cipher, _ := Encrypt(NewKey("secret-a"), "hello")
	// 翻转中间一个字节模拟篡改
	b := []byte(cipher)
	mid := len(b) / 2
	if b[mid] == 'A' {
		b[mid] = 'B'
	} else {
		b[mid] = 'A'
	}
	if _, err := Decrypt(NewKey("secret-a"), string(b)); err == nil {
		t.Fatal("expected error for tampered ciphertext")
	}
}

func TestDecryptGarbage(t *testing.T) {
	if _, err := Decrypt(NewKey("secret-a"), "not-base64!!"); err == nil {
		t.Fatal("expected error for garbage input")
	}
	if _, err := Decrypt(NewKey("secret-a"), "aGVsbG8="); err == nil {
		t.Fatal("expected error for too-short ciphertext")
	}
}
