# 🔐 Bank Informer

![Gopher](.github/gopher.webp)

Bank Informer is a Go-based application that retrieves and logs the balances of Ethereum Virtual Machine (EVM) and Pocket Network (POKT) wallets. It fetches the exchange rates for a list of currencies and calculates the fiat values for each balance. The balances, fiat values, and exchange rates are then logged.

Currently works for the following tokens: `USDC, USDT, ETH, POKT, WPOKT, WBTC`

**👋 Pull requests welcome to support additional ERC20 tokens.**

<p align="center">
<a href="https://github.com/buildwithgrove/path">
  <img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove Logo" title="Uses PATH for RPC">
  <br>
  🌿 Uses Grove's PATH (PATH API & Toolkit Harness) for RPC.
  <br>
</a>
</p>

## 🛠️ Configuration File

The application uses a YAML configuration file that lives at `$HOME/bank-informer/.bankinformer.config.yaml`. This file contains all the necessary settings for the PATH API & Toolkit Harness as well as wallet addresses and API keys.

The required configuration keys are:
- `path_api_url`: The URL for the PATH API & Toolkit Harness.
- `path_api_key`: Your PATH API KEY.
- `eth_wallet_address`: Your Ethereum wallet address.
- `pokt_wallet_address`: Your POKT wallet address.
- `cmc_api_key`: The CoinMarketCap API key used for fetching fiat-crypto exchange rates.

Optional configuration keys:
- `crypto_fiat_conversion`: The fiat currency to convert crypto balances to. Defaults to "USD".
- `convert_currencies`: A list of fiat currencies for which to fetch exchange rates. Defaults to ["USD"].
- `crypto_values`: A list of cryptocurrencies to display values for. Defaults to ["USDC", "ETH", "POKT"].

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
