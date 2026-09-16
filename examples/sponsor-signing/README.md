# Sponsor signing

This example signs a one-drop Payment in three modes:

- `cosigned`: the account signs, then a single sponsor signs.
- `multisigned`: the account signs, then two sponsor signer wallets independently sign the same account blob. The helper combines their fragments.
- `prefunded`: the account uses an existing ledger sponsorship without a sponsor signature.

## Requirements

Use a test node with the `Sponsor` amendment. Co-signing also requires `fixCleanup3_4_0`. The wallet helpers use the new sponsor signing prefixes and do not fall back to legacy prefixes.

Set these environment variables through your local secret manager or shell environment. Do not commit or log real seeds.

| Variable | Purpose |
| --- | --- |
| `XRPL_RPC_URL` | Compatible node URL |
| `SPONSOR_MODE` | `cosigned`, `multisigned`, or `prefunded` |
| `ACCOUNT_SEED` | Funded sending account's seed |
| `DESTINATION` | Existing destination account |
| `SPONSOR_ADDRESS` | Sponsor account's classic address |
| `SPONSOR_SEED` | Sponsor seed in `cosigned` mode, or first signer seed in `multisigned` mode |
| `SECOND_SPONSOR_SIGNER_SEED` | Second signer seed in `multisigned` mode |
| `SUBMIT` | Optional. Set to `1` to submit the final blob |

Run from the repository root:

```sh
go run ./examples/sponsor-signing
```

The default is to sign and print the final blob and hash, not submit it. This example does not fund accounts or set up signer lists or sponsorship ledger objects.

In `multisigned` mode, the two signer accounts must belong to the sponsor's signer list and meet its quorum. They must be distinct. For regular keys, use `SignAsSponsorOptions.MultisignAccount` to name the signer account rather than its key-derived address.

In `prefunded` mode, the existing `Sponsorship` object must have sufficient fee funds and allow fee use without a sponsor signature. The helper does not check these ledger conditions. Sponsorship does not pay the Payment amount.

## Signing order and fees

1. Set Sponsor and SponsorFlags on the unsigned transaction.
2. Call `AutofillMultisigned` with the total planned account and sponsor multisigner count. A single account or sponsor signature contributes zero extra signers.
3. Sign the account transaction. For account multisigning, combine account fragments before sponsor signing.
4. Sign each sponsor fragment independently, then combine them if needed.
5. Submit the final blob without another autofill or signing step.

Leave Fee absent before autofill. Autofill does not overwrite a supplied Fee. For an ordinary Payment, the signature fee factor is `1 + account multisigners + sponsor multisigners`. The existing fee engine handles transaction-specific costs and configured fee limits.

The helpers validate structure, not cryptographic validity of supplied signatures, ledger authorization, signer quorum, funds, or amendment activation. The combiner retains the first occurrence of a repeated decoded signer identity. This can retain an invalid signature even if a later fragment contains a valid one.
