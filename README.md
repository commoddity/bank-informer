# Bank Informer

![Gopher](.github/gopher.webp)

Bank Informer is a Go-based application that retrieves and logs the balances of Ethereum (ETH) and Pocket Network (POKT) wallets. It fetches the exchange rates for a list of currencies and calculates the fiat values for each balance. The balances, fiat values, and exchange rates are then logged.

Currently works for the following tokens: `USDC, ETH, POKT, WPOKT, WBTC`

**Pull requests welcome to support additional ERC20 tokens.**

<p align="center">
<a href="https://portal.grove.city/">
<img src=".github/grove_logo.png" alt="Grove Logo" title="Uses the Grove Portal for RPC">
<br>
Uses the Grove Portal for RPC.
<br>
</a>

</p>

## Environment Variables

The application requires several environment variables to function correctly. These are:

- `GROVE_PORTAL_APP_ID`: The application ID for the Grove portal, used for RPC.
- `ETH_WALLET_ADDRESS`: The Ethereum wallet address to fetch the balance from.
- `POKT_WALLET_ADDRESS`: The POKT wallet address to fetch the balance from.
- `CMC_API_KEY`: The API key for the CoinMarketCap service, used to fetch fiat-crypto exchange rates.

There are also several optional environment variables:

- `CRYPTO_FIAT_CONVERSION`: The fiat currency to convert the crypto balances to. Defaults to "USD".
- `CONVERT_CURRENCIES`: A comma-separated list of fiat currencies to fetch exchange rates for. Defaults to "USD".
- `CRYPTO_VALUES`: A comma-separated list of cryptocurrencies to display values for. Defaults to "USDC,ETH,POKT".

These environment variables can be set in a `.env.bankinformer` file in your home directory. To make changes to any of hte above variables just edit this file.

## Installation

Run
```bash
go install github.com/commoddity/bank-informer@latest
```

Then just run 
```bash
bank-informer
```

For the first run the application will ask you for the above environment variables and place them in a `.env.bankinformer` file in your home directory.

After that just run `bank-informer` to fetch your balances. 🚀
