## v0.4.1 - 2021-03-07

- Add sending lock/mutex (prevents `concurrent write to websocket connection` error)

## v0.4.0 - 2021-02-24

- Add direct API to worker messaging
- Add `redis_max_retries` option
- Add `redis_tls` option
- Added request ID to error responses
- Added more logging (including on error responses), and change multiple debug logging to info
- Parallelized parsing of messages received from Redis (should increase throughput)
- Deprecating `address` option: Use `port` instead
- Fix bug where creating a claim with ID left ID blank

## v0.3.3 - 2020-07-15

- Add `/ping` endpoint on API & worker

## v0.3.2 - 2020-07-12

- Rebuild with fixed Docker images

## v0.3.1 - 2020-06-14

- Improve query binding
- Add locking for concurrent state (thanks a lot [@abdullah-aghayan](https://github.com/abdullah-aghayan)!)
- Change logger to Zap (include request logger)
- Added logging to all requests
- Change unsubscribe type from 2 to 1 (internal type)
- Fix debug mode never being active
- Add log request (`log_requests`) option

## v0.3.0 - 2020-05-08

- **Breaking**
    - Change `MISSING_CONNECTION_OR_USER` to `MISSING_TARGET` (and changed message)
- Added channels:
    - Added `channel` target option to sending, getting info, disconnect
    - Added `channels` JWT claim (array of strings)
    - Added `channels` to claim creation (comma-delimited string)
    - Added `channels` to info response for connections
    - Added channel subscribing and unsubscribing
    - Added `default_channels` config options (comma-delimited string)
- Replace `scripts` directory with [Task](https://taskfile.dev)

## v0.2.0 - 2020-05-06

- **Breaking**
  - Changed response format for creating claims
    - All claim data is now inside the `claim` key, and more data is present. Example:
    
      Before:
      ```json
      {
        "success": true,
        "id": "XXX",
        "expiration": 1588473164
      }
      ```
      
      After:
      ```json
      {
        "success": true,
        "claim": {
          "id": "XXX",
          "expiration": 1588473164,
          "user": "a",
          "session": "b"
        }
      }
      ```

- Added tests (E2E, unit) and CI
- Added past expiration validation (doesn't allow expiration dates in the past) during claim creation

## v0.1.2 - 2020-05-06

- Improved code documentation, refactored to make it cleaner
- Improved API response code (refactored error handling)
- Small performance improvements for some hot-paths
- Actually bump version in code

## v0.1.1 - 2020-05-05

- Fixed bug with missing `errorCode` when duration is negative during claim creation
- Fix bug with incorrect `errorCode` when error is trigger during checking if a claim exists
- Added documentation on errors

> Version code during application startup is wrongly reported as v0.1.0

## v0.1.0 - 2020-05-04

- Initial beta release
