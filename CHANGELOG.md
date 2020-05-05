## Unreleased

- **Breaking**
  - Changed response format for creating claims
    - All claim data is inside the `claim` key, and more data is present
      Example:
      ```json5
      // Before
      {
        "success": true,
        "id": "XXX",
        "expiration": 1588473164
      }
      
      // After
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

- Added tests (E2E)

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
