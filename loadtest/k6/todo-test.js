/**
 * K6 Load Test - Todo CRUD Endpoints
 * Tests: Create, Get All, Get By ID, Update, Delete
 *
 * Run: k6 run loadtest/k6/todo-test.js
 * Run with options: k6 run --vus 50 --duration 2m loadtest/k6/todo-test.js
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { CONFIG, generateTestUser, generateTodo, getHeaders } from './config.js';

// Custom metrics for each operation
const createTodoSuccess = new Rate('create_todo_success');
const getTodosSuccess = new Rate('get_todos_success');
const getTodoByIdSuccess = new Rate('get_todo_by_id_success');
const updateTodoSuccess = new Rate('update_todo_success');
const deleteTodoSuccess = new Rate('delete_todo_success');

const createTodoDuration = new Trend('create_todo_duration', true);
const getTodosDuration = new Trend('get_todos_duration', true);
const getTodoByIdDuration = new Trend('get_todo_by_id_duration', true);
const updateTodoDuration = new Trend('update_todo_duration', true);
const deleteTodoDuration = new Trend('delete_todo_duration', true);

const todosCreated = new Counter('todos_created');
const todosDeleted = new Counter('todos_deleted');

// Test configuration
export const options = {
    scenarios: {
        // Main CRUD test scenario
        todo_crud: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 20 },   // Ramp up
                { duration: '2m', target: 20 },    // Sustained load
                { duration: '30s', target: 50 },   // Peak load
                { duration: '1m', target: 50 },    // Sustained peak
                { duration: '30s', target: 0 },    // Ramp down
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500', 'p(99)<1000'],
        http_req_failed: ['rate<0.02'],
        create_todo_success: ['rate>0.95'],
        get_todos_success: ['rate>0.98'],
        get_todo_by_id_success: ['rate>0.95'],
        update_todo_success: ['rate>0.95'],
        delete_todo_success: ['rate>0.95'],
    },
};

// Setup: Register user and get auth token
export function setup() {
    const user = generateTestUser();

    // Register user
    const registerRes = http.post(
        `${CONFIG.BASE_URL}/api/auth/register`,
        JSON.stringify(user),
        { headers: getHeaders() }
    );

    if (registerRes.status !== 201) {
        console.error(`Setup failed: Could not register user. Status: ${registerRes.status}`);
        console.error(`Response: ${registerRes.body}`);
        return { token: null, user: null };
    }

    const token = registerRes.json('data.token');
    console.log(`Setup: Registered user ${user.email}`);

    return { token, user };
}

// Main test function - Full CRUD cycle
export default function (data) {
    if (!data.token) {
        console.error('No auth token available. Skipping test.');
        return;
    }

    const headers = getHeaders(data.token);
    let createdTodoId = null;

    // ============================================
    // TEST 1: Create Todo
    // ============================================
    group('Create Todo', () => {
        const todo = generateTodo();

        const startTime = Date.now();
        const res = http.post(
            `${CONFIG.BASE_URL}/api/todos`,
            JSON.stringify(todo),
            { headers }
        );
        const duration = Date.now() - startTime;

        createTodoDuration.add(duration);

        const success = check(res, {
            'create todo status is 201': (r) => r.status === 201,
            'create todo returns id': (r) => {
                try {
                    const body = r.json();
                    if (body.data && body.data.id) {
                        createdTodoId = body.data.id;
                        return true;
                    }
                    return false;
                } catch {
                    return false;
                }
            },
            'create todo returns correct title': (r) => {
                try {
                    return r.json().data.title === todo.title;
                } catch {
                    return false;
                }
            },
        });

        createTodoSuccess.add(success);
        if (success) {
            todosCreated.add(1);
        }
    });

    sleep(0.5);

    // ============================================
    // TEST 2: Get All Todos (with pagination)
    // ============================================
    group('Get All Todos', () => {
        const startTime = Date.now();
        const res = http.get(
            `${CONFIG.BASE_URL}/api/todos?page=1&page_size=10`,
            { headers }
        );
        const duration = Date.now() - startTime;

        getTodosDuration.add(duration);

        const success = check(res, {
            'get todos status is 200': (r) => r.status === 200,
            'get todos returns array': (r) => {
                try {
                    const body = r.json();
                    return body.data && Array.isArray(body.data);
                } catch {
                    return false;
                }
            },
            'get todos has pagination info': (r) => {
                try {
                    const body = r.json();
                    return body.page !== undefined && body.total !== undefined;
                } catch {
                    return false;
                }
            },
        });

        getTodosSuccess.add(success);
    });

    sleep(0.5);

    // ============================================
    // TEST 3: Get Todo By ID
    // ============================================
    if (createdTodoId) {
        group('Get Todo By ID', () => {
            const startTime = Date.now();
            const res = http.get(
                `${CONFIG.BASE_URL}/api/todos/${createdTodoId}`,
                { headers }
            );
            const duration = Date.now() - startTime;

            getTodoByIdDuration.add(duration);

            const success = check(res, {
                'get todo by id status is 200': (r) => r.status === 200,
                'get todo by id returns correct id': (r) => {
                    try {
                        return r.json().data.id === createdTodoId;
                    } catch {
                        return false;
                    }
                },
            });

            getTodoByIdSuccess.add(success);
        });

        sleep(0.5);

        // ============================================
        // TEST 4: Update Todo
        // ============================================
        group('Update Todo', () => {
            const updateData = {
                title: `Updated Todo - ${Date.now()}`,
                completed: true,
            };

            const startTime = Date.now();
            const res = http.put(
                `${CONFIG.BASE_URL}/api/todos/${createdTodoId}`,
                JSON.stringify(updateData),
                { headers }
            );
            const duration = Date.now() - startTime;

            updateTodoDuration.add(duration);

            const success = check(res, {
                'update todo status is 200': (r) => r.status === 200,
                'update todo reflects changes': (r) => {
                    try {
                        const body = r.json();
                        return body.data.completed === true;
                    } catch {
                        return false;
                    }
                },
            });

            updateTodoSuccess.add(success);
        });

        sleep(0.5);

        // ============================================
        // TEST 5: Delete Todo
        // ============================================
        group('Delete Todo', () => {
            const startTime = Date.now();
            const res = http.del(
                `${CONFIG.BASE_URL}/api/todos/${createdTodoId}`,
                null,
                { headers }
            );
            const duration = Date.now() - startTime;

            deleteTodoDuration.add(duration);

            const success = check(res, {
                'delete todo status is 200': (r) => r.status === 200,
                'delete todo success message': (r) => {
                    try {
                        return r.json().success === true;
                    } catch {
                        return false;
                    }
                },
            });

            deleteTodoSuccess.add(success);
            if (success) {
                todosDeleted.add(1);
            }
        });

        // Verify deletion
        group('Verify Deletion', () => {
            const res = http.get(
                `${CONFIG.BASE_URL}/api/todos/${createdTodoId}`,
                { headers }
            );

            check(res, {
                'deleted todo returns 404': (r) => r.status === 404,
            });
        });
    }

    sleep(Math.random() * 2 + 1); // 1-3 seconds between iterations
}

// Teardown
export function teardown(data) {
    console.log('Todo CRUD Load Test completed');
}
