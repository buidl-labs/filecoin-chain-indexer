# filecoin-chain-indexer

Chain indexer for the Filecoin network

## Getting Started

- Clone the repo: `git clone https://github.com/buidl-labs/filecoin-chain-indexer` and `cd` into it.
- Build: `make build`
- Setup & run postgres server [instructions](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-18-04)
- Create db: `createdb filecoinminermarketplace`
- Migrate db: `./main --cmd=migrate`
- Set environment variables ([sample](.env.sample))

## List of Tasks

### 1. ActorCodesTask

- Invocation: `./main --cmd=actorcodes`
- Indexes _all_ actors (till a particular epoch) in the network and their corresponding actor code, stores them in the json file [ACTOR_CODES_JSON](actorCodes.json)
- The epoch till which actor codes have been indexed is stored in the file [ACS_PARSEDTILL](acspt)

### 2. MinersTask

- Invocation: `./main --cmd=miners`
- Indexes miner info (and funds) at [EPOCH](.env.sample) or (latest tipset - [900](https://spec.filecoin.io/#section-glossary.finality))
- Overwrites the values in the db forever
- If [ACTIVE_MINERS_ONLY](.env.sample) is `1` then just the miners specified in [ACTIVE_MINERS](servedir/activeMiners.json) will be indexed; else all miners in the network will be indexed

### 3. BlocksTask

- Invocation: `./main --cmd=blocks`
- Indexes all the block_headers from epoch `TO` through `FROM` (backward)

### 4. MessagesTask

- Invocation: `./main --cmd=messages`
- Indexes all transactions
  - from epoch `TO` through `FROM` (backward) if `FOREVER!=1`
  - from epoch `FROM` to [ACS_PARSEDTILL](acspt) (forward) if `FOREVER=1`. In this case the process runs forever since the `ACS_PARSEDTILL` value keeps increasing.
- During the transactions indexing process, all the changes in owner/worker/control addresses are parsed and written to [ADDR_CHANGES](servedir/addrChanges.json)

### 5. MarketsTask

- Invocation: `./main --cmd=markets`
- Fetches all ongoing market deals at [EPOCH](.env.sample) or (latest tipset - [900](https://spec.filecoin.io/#section-glossary.finality)) and inserts them to the db
