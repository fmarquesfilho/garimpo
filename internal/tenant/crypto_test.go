package tenant

import "testing"

func TestEncryptDecryptRoundtrip(t *testing.T) {
	cases := []string{
		"",
		"simple-secret",
		"MJS67QHU7HMCRX5AHI75YI2FO4M2AIXP",
		"123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		"special chars: ñ, ç, à, 日本語",
	}
	for _, plain := range cases {
		enc, err := Encrypt(plain)
		if err != nil {
			t.Fatalf("Encrypt(%q) falhou: %v", plain, err)
		}
		if plain == "" && enc == "" {
			continue // vazio não criptografa
		}
		if enc == plain {
			t.Errorf("Encrypt(%q) retornou plaintext inalterado", plain)
		}

		dec, err := Decrypt(enc)
		if err != nil {
			t.Fatalf("Decrypt falhou para Encrypt(%q): %v", plain, err)
		}
		if dec != plain {
			t.Errorf("Roundtrip falhou: original=%q, decrypted=%q", plain, dec)
		}
	}
}

func TestDecryptInvalid(t *testing.T) {
	_, err := Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("Decrypt de dados inválidos deveria retornar erro")
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	// Cada encrypt deve produzir nonce diferente
	enc1, _ := Encrypt("same-secret")
	enc2, _ := Encrypt("same-secret")
	if enc1 == enc2 {
		t.Error("Duas chamadas de Encrypt com mesmo input produziram mesmo ciphertext (nonce reutilizado?)")
	}
}
