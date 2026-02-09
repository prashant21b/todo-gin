# K6 Load Tests for Todo API

This directory contains load tests for the Todo API using [k6](https://k6.io/).

## Installation

### Install k6

**macOS:**
```bash
brew install k6
```

**Linux (Debian/Ubuntu):**
```bash
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

**Windows:**
```bash
choco install k6
```

**Docker:**
```bash
docker pull grafana/k6
```

## Test Files

| File | Description | Duration | Use Case |
|------|-------------|----------|----------|
| `quick-test.js` | Simple sanity check | ~30s | Quick verification |
| `auth-test.js` | Register & Login tests | ~4min | Auth endpoint testing |
| `todo-test.js` | Full CRUD operations | ~5min | Todo endpoint testing |
| `full-test.js` | Complete test suite | ~12min | Comprehensive testing |
| `spike-test.js` | Sudden traffic burst | ~4min | Resilience testing |

## Running Tests

### Prerequisites

1. Start your Todo API server:
```bash
go run ./cmd/server
```

2. Ensure the server is running on `http://localhost:8080` (default)

### Quick Test (Recommended for First Run)
```bash
k6 run loadtest/k6/quick-test.js
```

### Auth Endpoints Test
```bash
k6 run loadtest/k6/auth-test.js
```

### Todo CRUD Test
```bash
k6 run loadtest/k6/todo-test.js
```

### Full Test Suite
```bash
k6 run loadtest/k6/full-test.js
```

### Spike Test
```bash
k6 run loadtest/k6/spike-test.js
```

## Custom Options

### Change Base URL
```bash
k6 run -e BASE_URL=http://localhost:3000 loadtest/k6/quick-test.js
```

### Override VUs and Duration
```bash
k6 run --vus 20 --duration 1m loadtest/k6/quick-test.js
```

### Generate JSON Output
```bash
k6 run --out json=results.json loadtest/k6/full-test.js
```

### Generate HTML Report (using k6-reporter)
```bash
# Install k6-reporter first
k6 run loadtest/k6/full-test.js --out json=results.json

# Then convert to HTML (requires external tool)
```

## Understanding Results

### Key Metrics

| Metric | Description | Good Value |
|--------|-------------|------------|
| `http_req_duration` | Response time | p95 < 500ms |
| `http_req_failed` | Failed requests | < 1% |
| `http_reqs` | Requests per second | Higher = better |
| `vus` | Virtual users | Depends on scenario |

### Reading the Output

```
     ✓ create: status 201
     ✓ list: status 200
     ✓ get: status 200
     ✓ update: status 200
     ✓ delete: status 200

     checks.....................: 100.00% ✓ 500  ✗ 0
     http_req_duration..........: avg=45ms  min=12ms  p(90)=89ms  p(95)=120ms
     http_req_failed............: 0.00%   ✓ 0    ✗ 500
     http_reqs..................: 500     16.5/s
```

- ✓ = Passed checks
- ✗ = Failed checks
- `p(95)` = 95th percentile (95% of requests were faster than this)

## Test Scenarios Explained

### 1. Smoke Test
- **Purpose**: Basic functionality verification
- **Load**: 1 user
- **Duration**: 30 seconds
- **Use**: After deployments, quick sanity checks

### 2. Load Test
- **Purpose**: Test normal expected load
- **Load**: 20 concurrent users
- **Duration**: 5 minutes
- **Use**: Regular performance validation

### 3. Stress Test
- **Purpose**: Find system limits
- **Load**: Up to 100 users
- **Duration**: 7 minutes
- **Use**: Capacity planning

### 4. Spike Test
- **Purpose**: Test sudden traffic bursts
- **Load**: 10 → 100 → 10 users rapidly
- **Duration**: 4 minutes
- **Use**: Resilience testing

## Troubleshooting

### Connection Refused
```
ERRO[0001] dial tcp 127.0.0.1:8080: connect: connection refused
```
**Solution**: Make sure your API server is running.

### Too Many Open Files
```
WARN[0030] Request Failed error="dial tcp: lookup localhost: too many open files"
```
**Solution**: Increase file descriptor limit:
```bash
ulimit -n 10000
```

### High Error Rate
If you see high error rates, check:
1. Database connection pool size
2. Server resource limits
3. Network bandwidth

## Customizing Tests

### Adding New Endpoints

Edit the test files and add new functions:

```javascript
function myNewEndpoint(token) {
    const res = http.get(
        `${CONFIG.BASE_URL}/api/my-endpoint`,
        { headers: getHeaders(token) }
    );

    check(res, {
        'my endpoint: status 200': (r) => r.status === 200,
    });
}
```

### Changing Thresholds

Edit `options.thresholds` in any test file:

```javascript
export const options = {
    thresholds: {
        http_req_duration: ['p(95)<300'],  // Stricter: 95% under 300ms
        http_req_failed: ['rate<0.001'],   // Stricter: less than 0.1% failures
    },
};
```

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Run Load Tests
  run: |
    k6 run loadtest/k6/quick-test.js --out json=k6-results.json

- name: Upload Results
  uses: actions/upload-artifact@v3
  with:
    name: k6-results
    path: k6-results.json
```

## Resources

- [k6 Documentation](https://k6.io/docs/)
- [k6 Examples](https://github.com/grafana/k6/tree/master/examples)
- [k6 Best Practices](https://k6.io/docs/testing-guides/api-load-testing/)
