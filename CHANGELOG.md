## Unreleased

- Improve query binding
- Add locking for concurrent state (thanks a lot @abdullah-aghayan!)

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
