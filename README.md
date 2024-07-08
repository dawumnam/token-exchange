# Token Trader

## How to get started on MacOS
1. Make sure golang, brew and docker is installed
2. `git clone https://github.com/dawumnam/token-exchange.git`
3. `cd token-exchange`
4. `brew install golang-migrate`
5. `docker compose up -d`
6. `make migrate-up`
7. `make test`
8. `make run`

## Features Implemented
- User registration, authentication, and logout
- Token creation and deployment (onchain)
- Token balance checking (offchain)
- Token transfer between users (offchain)
- Order List checking (offchain)

## Other Implementations
- DB and Cache dockerization
- E2e and unit tests

## Features not implemented 
- Logging to a file
- Stress testing

## TBD
- Scheduled settlements (apply offchain transfers to onchain)
- Using some form of authentic currency as a payment source for trades
- Many more.. 
