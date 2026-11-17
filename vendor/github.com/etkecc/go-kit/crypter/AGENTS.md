# crypter

AES-GCM string encryption with the `ENCv1[...]` tag; root module, thread-safe, and structurally usable by the YAML walker.

## Contract

- `New` accepts exactly 16, 24, or 32 raw key bytes (AES-128/192/256); the secret is used directly, with no KDF or salt.
- `Encrypt` is idempotent for tagged input; `Decrypt` passes untagged input through unchanged.
- Format is `ENCv1[base64url-raw(nonce || ciphertext || tag)]`; a 12-byte random nonce per call; one `Crypter` serves many goroutines.
- Quote encrypted values in YAML/JSON so the trailing `]` is not parsed as a flow sequence.

## Landmines

- Plaintext must never start with `ENCv1[`; `Encrypt` treats it as already sealed and stores it in the clear.
- A tampered value or the wrong key returns `ErrOpen`; treat that as terminal and discard the ciphertext.
- The secret is raw key material; a memorable passphrase fails the length check.

## Verify

`go test ./crypter`

## Pointers

- `README.md` - tag format, key handling, error behavior.
- Root rules: `../AGENTS.md`.
