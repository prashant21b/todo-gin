// Load test configuration
export const CONFIG = {
    BASE_URL: __ENV.BASE_URL || 'http://localhost:8080',

    // Thresholds for pass/fail criteria
    THRESHOLDS: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],  // 95% < 500ms, 99% < 1s
        http_req_failed: ['rate<0.01'],                   // Error rate < 1%
    },
};

// Generate unique test user for each test run
export function generateTestUser() {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substring(7);
    return {
        name: `Load Test User ${random}`,
        email: `loadtest_${timestamp}_${random}@example.com`,
        password: 'Password123!',
    };
}

// Common headers helper
export function getHeaders(token = null) {
    const headers = {
        'Content-Type': 'application/json',
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    return headers;
}

// Random todo generator
export function generateTodo() {
    const titles = [
        'Complete project documentation',
        'Review pull requests',
        'Fix bug in authentication',
        'Write unit tests',
        'Deploy to staging',
        'Update dependencies',
        'Refactor database queries',
        'Implement caching',
        'Add error handling',
        'Optimize performance',
    ];

    const categories = ['Work', 'Personal', 'Shopping', 'Health', 'Learning'];

    return {
        title: titles[Math.floor(Math.random() * titles.length)] + ` - ${Date.now()}`,
        description: `Task created during load test at ${new Date().toISOString()}`,
        category: categories[Math.floor(Math.random() * categories.length)],
    };
}
