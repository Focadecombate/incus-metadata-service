# Database Mocks

This package contains reusable mocks for the database layer.

## Usage

To use the MockQuerier in your tests:

```go
package your_package_test

import (
    "testing"
    "github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db/mocks"
    "github.com/stretchr/testify/mock"
)

func TestYourFunction(t *testing.T) {
    // Create a new mock instance
    mockDB := &mocks.MockQuerier{}
    
    // Setup expectations
    mockDB.On("GetInstance", mock.Anything, mock.Anything).Return(db.Instance{}, nil)
    
    // Use the mock in your handler/service
    handler := &YourHandler{
        Database: mockDB,
    }
    
    // Run your test logic
    // ...
    
    // Verify expectations were met
    mockDB.AssertExpectations(t)
}
```

## Available Methods

The MockQuerier implements all methods from the `db.Querier` interface:

### Vendor Data

- `CreateVendorData`
- `GetVendorData`
- `UpdateVendorData`
- `DeleteVendorData`

### Instances

- `CreateInstance`
- `GetInstance`
- `GetInstanceByID`
- `GetInstanceByIP`
- `ListInstances`
- `ListInstancesByProject`
- `UpdateInstance`
- `UpdateInstanceIP`
- `DeleteInstance`
- `HardDeleteInstance`

### Instance State

- `CreateOrUpdateInstanceState`
- `GetInstanceState`
- `DeleteInstanceState`

### Instance Logs

- `CreateInstanceLog`
- `GetInstanceLogs`
- `GetInstanceLogsByType`
- `GetInstanceLogsByLevel`
- `DeleteInstanceLogs`
- `DeleteOldInstanceLogs`

### Profiles

- `CreateProfile`
- `GetProfile`
- `ListProfiles`
- `ListProfilesByProject`
- `UpdateProfile`
- `DeleteProfile`
