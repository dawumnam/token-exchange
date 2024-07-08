# Token Trader

## How to get started on MacOS
1. Make sure golang, brew and docker is installed
2. `brew install golang-migrate`
3. `docker compose up -d`
4. `make migrate-up`
5. `make test`
6. `make run`

## Features Implemented
- User registration and authentication
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
