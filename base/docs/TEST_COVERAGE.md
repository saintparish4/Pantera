# Test Coverage Report

## Service Package

### Test Results

```bash
$ go test -cover ./internal/service
ok      github.com/saintparish4/harmonia/internal/service     1.809s  coverage: 36.9% of statements
```

### Coverage Summary

- **Coverage**: 36.9% of statements
- **Test Duration**: 1.809s
- **Package**: `github.com/saintparish4/harmonia/internal/service`
- **Status**: All tests passed (ok)

### Test Files

- `pricing_engine_test.go` - Tests for PricingEngine
- `cost_plus_test.go` - Tests for CostPlusStrategy
- `geographic_test.go` - Tests for GeographicStrategy
- `time_based_test.go` - Tests for TimeBasedStrategy
- `rule_based_test.go` - Tests for RuleBasedStrategy

## Repository Package

### Test Results

```bash
$ go test -cover ./internal/repository
ok      github.com/saintparish4/harmonia/internal/repository  1.657s  coverage: 6.4% of statements
```

### Coverage Summary

- **Coverage**: 6.4% of statements
- **Test Duration**: 1.657s
- **Package**: `github.com/saintparish4/harmonia/internal/repository`
- **Status**: All tests passed (ok)

### Test Files

- `api_key_utils_test.go` - Tests for API key utilities (generation, validation, masking, hashing)

## Running Tests

To run tests with coverage for a specific package:
```bash
go test -cover ./internal/service
go test -cover ./internal/repository
```

To run all tests:
```bash
go test -cover ./...
```

To generate detailed coverage report:
```bash
go test -coverprofile=coverage.out ./internal/service
go tool cover -html=coverage.out
```

