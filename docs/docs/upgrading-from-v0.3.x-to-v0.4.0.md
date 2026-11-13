# Upgrade from v0.3.x to v0.4.0

This guide covers the source and behavior changes that are most likely to affect applications that upgrade from `v0.3.x` to `v0.4.0`. See the [changelog](https://github.com/XRPLF/xrpl-go/blob/main/CHANGELOG.md) for the complete list.

## Review runtime behavior

### Wallet mnemonics

If your original mnemonic has one space between each word and no surrounding whitespace, and any of those spaces are non-ASCII spaces that Unicode NFKD maps to ASCII spaces, such as non-breaking spaces, `wallet.FromMnemonic` now derives the address of the same words separated by single ASCII spaces, not the original address. It returns no error. Compare the derived address with your funded address before you use the wallet.

`FromMnemonic` now applies NFKD normalization before validation and derivation, like xrpl.js. Released versions `v0.1.0` through `v0.3.1` validated the words but derived the wallet from the original string. A newline, tab, repeated space, or leading or trailing space now returns an error. A mnemonic with single ASCII spaces between words and no leading or trailing whitespace is not affected by this change.

To check whether you are affected, compare your funded address with the address derived by the new version from your cleaned mnemonic:

```go
cleaned := strings.Join(strings.Fields(originalMnemonic), " ")
cleanedWallet, err := wallet.FromMnemonic(cleaned)
if err != nil {
	return err
}
fmt.Println("Cleaned-mnemonic address:", cleanedWallet.ClassicAddress)
```

Import `fmt`, `strings`, and `github.com/Peersyst/xrpl-go/xrpl/wallet` for this example. If the addresses differ, the cleaned-mnemonic wallet does not control the funded address by default. Keep the exact original string, including all whitespace. Never share the mnemonic or private keys.

If funds are on the old address, use `wallet.FromMnemonicLegacy` with the exact original string. It derives the wallet with the `FromMnemonic` behavior of `v0.1.13` through `v0.3.1`, so you can sign the migration transactions in the same program:

```go
original, err := os.ReadFile("mnemonic.txt") // exact original bytes, including any final newline
if err != nil {
	return err
}
oldWallet, err := wallet.FromMnemonicLegacy(string(original))
if err != nil {
	return err
}
fmt.Println("Old address:", oldWallet.ClassicAddress)
```

Also import `os` for this example.

Confirm the old address matches the funded address. Then use `oldWallet` to send the funds to the cleaned-mnemonic address with a `Payment`, or close the old account with `AccountDelete` if its ledger requirements permit. Alternatively, use `SetRegularKey` signed by `oldWallet` to authorize the cleaned-mnemonic wallet as the old account's regular key. Check the transaction result before discarding any recovery material. Test the process with non-production funds first. Keep the mnemonic and any private key private.

`FromMnemonicLegacy` is deprecated and exists only for this migration. It is available in all `v0.4.x` releases and can be removed in `v0.5.0` or any later release, so complete the migration before you upgrade to `v0.5.0`.

#### Wallets created with v0.1.0 through v0.1.12

For a small share of mnemonics, `v0.1.0` through `v0.1.12` derive a different address than later versions, even without whitespace. `FromMnemonicLegacy` reproduces `v0.1.13` through `v0.3.1` only. If its address does not match the funded address and the wallet was created with `v0.1.12` or earlier, recover it in a separate program pinned to `v0.1.12`:

```sh
mkdir mnemonic-recovery && cd mnemonic-recovery
go mod init mnemonic-recovery
go get github.com/Peersyst/xrpl-go/xrpl/wallet@v0.1.12
```

In that program, call `wallet.FromMnemonic(string(original))` with the same exact original string, confirm the address, and sign the same migration transactions.
