# BlackHole Blockchain API Specifications\n\n## Cryptography\n- ECDSA for digital signatures\n- Optional ZKPs for privacy\n\n## APIs\n- Relay Chain: /submit-transaction\n- Token/Wallet: /create-token, /transfer-token\n\n## TokenX\n- Native token for staking, governance

## Staking APIs
- POST /stake: Stake TokenX (address, target, amount, stakeType)
- POST /unstake: Unstake TokenX
- GET /claim-rewards: Claim staking rewards