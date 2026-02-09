/**
 * K6 Quick Load Test - Simple and Fast
 * Use this for quick sanity checks
 *
 * Run: k6 run loadtest/k6/quick-test.js
 * Run with more VUs: k6 run --vus 10 --duration 30s loadtest/k6/quick-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { CONFIG, generateTestUser, generateTodo, getHeaders } from './config.js';

// Simple configuration
export const options = {
    vus: 5,              // 5 virtual users
    duration: '30s',     // Run for 30 seconds
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.05'],
    },
};

// Setup - Create one user for all VUs
export function setup() {
    const user = generateTestUser();

    const res = http.post(
        `${CONFIG.BASE_URL}/api/auth/register`,
        JSON.stringify(user),
        { headers: getHeaders() }
    );

    if (res.status !== 201) {
        console.error('Setup failed:', res.body);
        return { token: null };
    }

    const token = res.json('data.token');
    console.log(`Setup complete. User: ${user.email}`);

    return { token, user };
}

// Main test
export default function (data) {
    if (!data.token) {
        console.error('No token available');
        return;
    }

    const headers = getHeaders(data.token);

    // 1. Create a todo
    const todo = generateTodo();
    const createRes = http.post(
        `${CONFIG.BASE_URL}/api/todos`,
        JSON.stringify(todo),
        { headers }
    );

    const created = check(createRes, {
        'create: status 201': (r) => r.status === 201,
    });

    if (!created) {
        sleep(1);
        return;
    }

    const todoId = createRes.json('data.id');

    // 2. Get all todos
    const listRes = http.get(`${CONFIG.BASE_URL}/api/todos`, { headers });
    check(listRes, {
        'list: status 200': (r) => r.status === 200,
    });

    // 3. Get single todo
    const getRes = http.get(`${CONFIG.BASE_URL}/api/todos/${todoId}`, { headers });
    check(getRes, {
        'get: status 200': (r) => r.status === 200,
    });

    // 4. Update todo
    const updateRes = http.put(
        `${CONFIG.BASE_URL}/api/todos/${todoId}`,
        JSON.stringify({ completed: true }),
        { headers }
    );
    check(updateRes, {
        'update: status 200': (r) => r.status === 200,
    });

    // 5. Delete todo
    const deleteRes = http.del(`${CONFIG.BASE_URL}/api/todos/${todoId}`, null, { headers });
    check(deleteRes, {
        'delete: status 200': (r) => r.status === 200,
    });

    sleep(1);
}

export function teardown(data) {
    console.log('Quick test completed!');
}
