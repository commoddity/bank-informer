# 🔐 Bank Informer

![Gopher](.github/gopher.png)

Bank Informer is a Go-based application that retrieves and logs the balances of Ethereum Virtual Machine (EVM) and Pocket Network (POKT) wallets. It fetches the exchange rates for a list of currencies and calculates the fiat values for each balance. The balances, fiat values, and exchange rates are then logged.

Currently works for the following tokens: `USDC, USDT, ETH, POKT, WPOKT, WBTC`

**👋 Pull requests welcome to support additional ERC20 tokens.**

<p align="center">
<a href="https://pocket.network">
  <img style="width: 400px;" src="https://pocket.network/wp-content/uploads/2025/06/Pocket-Logo-White.svg" alt="Pocket Network Logo" title="Uses Pocket Network for RPC">
  <br>
  🌐 Uses Pocket Network's free API endpoints for RPC.
  <br>
</a>
</p>

## 🛠️ Configuration File

The application uses a YAML configuration file that lives at `$HOME/bank-informer/.bankinformer.config.yaml`. This file contains all the necessary settings for wallet addresses and API keys.

The required configuration keys are:
- `eth_wallet_address`: Your Ethereum wallet address.
- `pokt_wallet_address`: Your POKT wallet address.
- `cmc_api_key`: The CoinMarketCap API key used for fetching fiat-crypto exchange rates.

Optional configuration keys:
- `rpc_api_key`: Your RPC API key for Pocket Network API authentication (optional).
- `crypto_fiat_conversion`: The fiat currency to convert crypto balances to. Defaults to "USD".
- `convert_currencies`: A list of fiat currencies for which to fetch exchange rates. Defaults to ["USD"].
- `crypto_values`: A list of cryptocurrencies to display values for. Defaults to ["USDC", "ETH", "POKT"].

The application uses Pocket Network's free API endpoints at `https://api.pocket.network` for RPC calls. Each package (ETH and POKT) automatically constructs the appropriate subdomain-based URL (e.g., `https://eth.api.pocket.network` or `https://pocket.api.pocket.network`).

By default, the YAML configuration file is created at `$HOME/bank-informer/.bankinformer.config.yaml`. You can edit this file at any time to update your configuration.

## 💻 Installation

Run:
```bash
go install github.com/commoddity/bank-informer@latest
```

Then, simply run:
```bash
bank-informer
```

For the first run, the application will prompt you to enter the required configuration values and will create the YAML configuration file automatically. After that, just run `bank-informer` to fetch your balances. 🚀
