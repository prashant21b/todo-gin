/**
 * K6 Load Test - Full API Test
 * Comprehensive test covering all endpoints with realistic user flows
 *
 * Run: k6 run loadtest/k6/full-test.js
 * Run with HTML report: k6 run loadtest/k6/full-test.js --out json=results.json
 */

import http from 'k6/http';
import { check, group, sleep, fail } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';
import { CONFIG, generateTestUser, generateTodo, getHeaders } from './config.js';

// ============================================
// Custom Metrics
// ============================================

// Success rates
const authSuccess = new Rate('auth_success');
const crudSuccess = new Rate('crud_success');
const overallSuccess = new Rate('overall_success');

// Duration trends
const authDuration = new Trend('auth_duration', true);
const crudDuration = new Trend('crud_duration', true);

// Counters
const totalRequests = new Counter('total_requests');
const failedRequests = new Counter('failed_requests');

// ============================================
// Test Configuration
// ============================================
export const options = {
    scenarios: {
        // Smoke test - Quick sanity check
        smoke_test: {
            executor: 'constant-vus',
            vus: 1,
            duration: '30s',
            exec: 'smokeTest',
            tags: { test_type: 'smoke' },
        },

        // Load test - Normal expected load
        load_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 20 },   // Ramp up
                { duration: '3m', target: 20 },   // Steady state
                { duration: '1m', target: 0 },    // Ramp down
            ],
            startTime: '30s',  // Start after smoke test
            exec: 'loadTest',
            tags: { test_type: 'load' },
        },

        // Stress test - Find breaking point
        stress_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '1m', target: 50 },   // Ramp up
                { duration: '2m', target: 50 },   // Stay at 50
                { duration: '1m', target: 100 },  // Push to 100
                { duration: '2m', target: 100 },  // Stay at 100
                { duration: '1m', target: 0 },    // Ramp down
            ],
            startTime: '5m30s',  // Start after load test
            exec: 'stressTest',
            tags: { test_type: 'stress' },
        },
    },

    thresholds: {
        // Global thresholds
        http_req_duration: ['p(95)<500', 'p(99)<1500'],
        http_req_failed: ['rate<0.05'],

        // Custom thresholds
        auth_success: ['rate>0.95'],
        crud_success: ['rate>0.90'],
        overall_success: ['rate>0.90'],

        // Per-scenario thresholds
        'http_req_duration{test_type:smoke}': ['p(95)<300'],
        'http_req_duration{test_type:load}': ['p(95)<500'],
        'http_req_duration{test_type:stress}': ['p(95)<1000'],
    },
};

// ============================================
// Helper Functions
// ============================================

function registerUser() {
    const user = generateTestUser();

    const res = http.post(
        `${CONFIG.BASE_URL}/api/auth/register`,
        JSON.stringify(user),
        { headers: getHeaders(), tags: { name: 'Register' } }
    );

    totalRequests.add(1);

    const success = check(res, {
        'register: status 201': (r) => r.status === 201,
        'register: has token': (r) => r.json('data.token') !== undefined,
    });

    if (!success) {
        failedRequests.add(1);
        return null;
    }

    return {
        user,
        token: res.json('data.token'),
    };
}

function loginUser(email, password) {
    const res = http.post(
        `${CONFIG.BASE_URL}/api/auth/login`,
        JSON.stringify({ email, password }),
        { headers: getHeaders(), tags: { name: 'Login' } }
    );

    totalRequests.add(1);

    const success = check(res, {
        'login: status 200': (r) => r.status === 200,
        'login: has token': (r) => r.json('data.token') !== undefined,
    });

    if (!success) {
        failedRequests.add(1);
        return null;
    }

    return res.json('data.token');
}

function createTodo(token) {
    const todo = generateTodo();

    const res = http.post(
        `${CONFIG.BASE_URL}/api/todos`,
        JSON.stringify(todo),
        { headers: getHeaders(token), tags: { name: 'CreateTodo' } }
    );

    totalRequests.add(1);

    const success = check(res, {
        'create todo: status 201': (r) => r.status === 201,
        'create todo: has id': (r) => r.json('data.id') !== undefined,
    });

    if (!success) {
        failedRequests.add(1);
        return null;
    }

    return res.json('data.id');
}

function getTodos(token, page = 1, pageSize = 10) {
    const res = http.get(
        `${CONFIG.BASE_URL}/api/todos?page=${page}&page_size=${pageSize}`,
        { headers: getHeaders(token), tags: { name: 'GetTodos' } }
    );

    totalRequests.add(1);

    const success = check(res, {
        'get todos: status 200': (r) => r.status === 200,
        'get todos: has data array': (r) => Array.isArray(r.json('data')),
    });

    if (!success) {
        failedRequests.add(1);
    }

    return success;
}

function getTodoById(token, todoId) {
    const res = http.get(
        `${CONFIG.BASE_URL}/api/todos/${todoId}`,
        { headers: getHeaders(token), tags: { name: 'GetTodoById' } }
    );

    totalRequests.add(1);

    const success = check(res, {
        'get todo by id: status 200': (r) => r.status === 200,
        'get todo by id: correct id': (r) => r.json('data.id') === todoId,
    });

    if (!success) {
        failedRequests.add(1);
    }

    return success;
}

function updateTodo(token, todoId) {
    const updateData = {
        title: `Updated at ${new Date().toISOString()}`,
        completed: Math.random() > 0.5,
    };

    const res = http.put(
        `${CONFIG.BASE_URL}/api/todos/${todoId}`,
        JSON.stringify(updateData),
        { headers: getHeaders(token), tags: { name: 'UpdateTodo' } }
    );

    totalRequests.add(1);

    const success = check(res, {
        'update todo: status 200': (r) => r.status === 200,
        'update todo: success true': (r) => r.json('success') === true,
    });

    if (!success) {
        failedRequests.add(1);
    }

    return success;
}

function deleteTodo(token, todoId) {
    const res = http.del(
        `${CONFIG.BASE_URL}/api/todos/${todoId}`,
        null,
        { headers: getHeaders(token), tags: { name: 'DeleteTodo' } }
    );

    totalRequests.add(1);

    const success = check(res, {
        'delete todo: status 200': (r) => r.status === 200,
        'delete todo: success true': (r) => r.json('success') === true,
    });

    if (!success) {
        failedRequests.add(1);
    }

    return success;
}

// ============================================
// Test Scenarios
// ============================================

// Smoke Test - Basic functionality check
export function smokeTest() {
    group('Smoke Test - Basic Flow', () => {
        // 1. Health check
        const healthRes = http.get(`${CONFIG.BASE_URL}/api/health`);
        check(healthRes, {
            'health check: status 200': (r) => r.status === 200,
        });

        // 2. Register
        const authData = registerUser();
        if (!authData) {
            fail('Smoke test failed: Could not register user');
            return;
        }

        // 3. Login
        const token = loginUser(authData.user.email, authData.user.password);
        if (!token) {
            fail('Smoke test failed: Could not login');
            return;
        }

        // 4. Create todo
        const todoId = createTodo(token);
        if (!todoId) {
            fail('Smoke test failed: Could not create todo');
            return;
        }

        // 5. Get todos
        getTodos(token);

        // 6. Get todo by ID
        getTodoById(token, todoId);

        // 7. Update todo
        updateTodo(token, todoId);

        // 8. Delete todo
        deleteTodo(token, todoId);

        overallSuccess.add(1);
    });

    sleep(1);
}

// Load Test - Normal user behavior simulation
export function loadTest() {
    const startTime = Date.now();

    group('Load Test - User Journey', () => {
        // Register new user
        const authData = registerUser();
        if (!authData) {
            authSuccess.add(0);
            overallSuccess.add(0);
            return;
        }
        authSuccess.add(1);

        const token = authData.token;
        sleep(0.5);

        // Create 3-5 todos
        const todoIds = [];
        const todoCount = Math.floor(Math.random() * 3) + 3;

        for (let i = 0; i < todoCount; i++) {
            const todoId = createTodo(token);
            if (todoId) {
                todoIds.push(todoId);
                crudSuccess.add(1);
            } else {
                crudSuccess.add(0);
            }
            sleep(0.2);
        }

        // Get all todos
        if (getTodos(token)) {
            crudSuccess.add(1);
        } else {
            crudSuccess.add(0);
        }
        sleep(0.3);

        // Get individual todos
        for (const todoId of todoIds) {
            if (getTodoById(token, todoId)) {
                crudSuccess.add(1);
            } else {
                crudSuccess.add(0);
            }
            sleep(0.1);
        }

        // Update some todos
        const todosToUpdate = todoIds.slice(0, Math.ceil(todoIds.length / 2));
        for (const todoId of todosToUpdate) {
            if (updateTodo(token, todoId)) {
                crudSuccess.add(1);
            } else {
                crudSuccess.add(0);
            }
            sleep(0.2);
        }

        // Delete one todo
        if (todoIds.length > 0) {
            if (deleteTodo(token, todoIds[0])) {
                crudSuccess.add(1);
            } else {
                crudSuccess.add(0);
            }
        }

        overallSuccess.add(1);
    });

    const duration = Date.now() - startTime;
    crudDuration.add(duration);

    sleep(Math.random() * 2 + 1);
}

// Stress Test - High load simulation
export function stressTest() {
    const startTime = Date.now();

    group('Stress Test - Rapid Operations', () => {
        // Register
        const authData = registerUser();
        if (!authData) {
            overallSuccess.add(0);
            return;
        }

        const token = authData.token;

        // Rapid create-read-delete cycles
        for (let i = 0; i < 5; i++) {
            const todoId = createTodo(token);
            if (todoId) {
                getTodoById(token, todoId);
                updateTodo(token, todoId);
                deleteTodo(token, todoId);
            }
            sleep(0.1);
        }

        // Multiple concurrent reads
        getTodos(token, 1, 20);
        getTodos(token, 2, 20);

        overallSuccess.add(1);
    });

    const duration = Date.now() - startTime;
    crudDuration.add(duration);

    sleep(0.5);
}

// Default function
export default function () {
    loadTest();
}

// Teardown
export function teardown(data) {
    console.log('\n========================================');
    console.log('Full Load Test Completed');
    console.log('========================================\n');
}
