# Bank Informer

![Gopher](.github/gopher.webp)

Bank Informer is a Go-based application that retrieves and logs the balances of Ethereum (ETH) and Pocket Network (POKT) wallets. It fetches the exchange rates for a list of currencies and calculates the fiat values for each balance. The balances, fiat values, and exchange rates are then logged.

Current works for the following tokens:
`USDC, ETH, POKT`

Pull requests welcome to support additional ERC20 tokens.

## Environment Variables

The application requires several environment variables to function correctly. These are:

- `ETH_WALLET_ADDRESS`: The Ethereum wallet address to fetch the balance from.
- `GROVE_PORTAL_APP_ID`: The application ID for the Pocket Network portal, used to fetch POKT wallet balances.
- `POKT_WALLET_ADDRESS`: The POKT wallet address to fetch the balance from.
- `CMC_API_KEY`: The API key for the CoinMarketCap service, used to fetch exchange rates.

There are also several optional environment variables:

- `CRYPTO_FIAT_CONVERSION`: The fiat currency to convert the crypto balances to. Defaults to "CAD".
- `CONVERT_CURRENCIES`: A comma-separated list of fiat currencies to fetch exchange rates for. Defaults to "CAD".
- `CRYPTO_VALUES`: A comma-separated list of cryptocurrencies to display values for. Defaults to "USDC,ETH,POKT".

These environment variables can be set in a `.env` file in the root directory of the project. This file is ignored by Git.

## Build and Run

To build and run the Bank Informer application, follow these steps:

1. **Build the Binary**: Depending on your operating system, use one of the following make commands to build the binary:

   - For Windows: `make build-windows`
   - For Linux: `make build-linux`
   - For macOS: `make build-mac`

   This will create a binary in the `bin` directory.

2. **Create a .env File**: In the root directory of the project, create a `.env` file to store your environment variables. Here's an example of what this file might look like:

```bash
ETH_WALLET_ADDRESS=your_ethereum_wallet_address
GROVE_PORTAL_APP_ID=your_GROVE_PORTAL_APP_ID
POKT_WALLET_ADDRESS=your_pokt_wallet_address
CMC_API_KEY=your_coinmarketcap_api_key

CRYPTO_FIAT_CONVERSION=CAD,EUR
CONVERT_CURRENCIES=CAD
CRYPTO_VALUES=USDC,ETH,POKT
```

To execute just place the `main` binary file (or `main.exe` for Windows) and `.env` in the same directory then run `./main` from the command line.

The application will now run.
