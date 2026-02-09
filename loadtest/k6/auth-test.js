/**
 * K6 Load Test - Authentication Endpoints
 * Tests: Register and Login
 *
 * Run: k6 run loadtest/k6/auth-test.js
 * Run with custom URL: k6 run -e BASE_URL=http://localhost:8080 loadtest/k6/auth-test.js
 */

import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { CONFIG, generateTestUser, getHeaders } from './config.js';

// Custom metrics
const registerSuccessRate = new Rate('register_success');
const loginSuccessRate = new Rate('login_success');
const registerDuration = new Trend('register_duration', true);
const loginDuration = new Trend('login_duration', true);
const registeredUsers = new Counter('registered_users');
const successfulLogins = new Counter('successful_logins');

// Test configuration
export const options = {
    scenarios: {
        // Scenario 1: Registration load test
        registration_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 10 },  // Ramp up to 10 users
                { duration: '1m', target: 10 },   // Stay at 10 users
                { duration: '30s', target: 0 },   // Ramp down
            ],
            exec: 'registerScenario',
        },
        // Scenario 2: Login load test (starts after registration)
        login_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 20 },  // Ramp up to 20 users
                { duration: '1m', target: 20 },   // Stay at 20 users
                { duration: '30s', target: 0 },   // Ramp down
            ],
            startTime: '2m',  // Start after registration test
            exec: 'loginScenario',
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500'],
        http_req_failed: ['rate<0.05'],
        register_success: ['rate>0.95'],
        login_success: ['rate>0.95'],
    },
};

// Note: Each VU gets its own copy of setup() data via the function parameter

// Setup: Create a user for login tests
export function setup() {
    const user = generateTestUser();

    const registerRes = http.post(
        `${CONFIG.BASE_URL}/api/auth/register`,
        JSON.stringify(user),
        { headers: getHeaders() }
    );

    if (registerRes.status === 201) {
        console.log(`Setup: Created test user ${user.email}`);
        return { user };
    } else {
        console.log(`Setup: Failed to create test user. Status: ${registerRes.status}`);
        return { user };
    }
}

// Scenario 1: Registration Test
export function registerScenario() {
    group('User Registration', () => {
        const user = generateTestUser();

        const startTime = Date.now();
        const res = http.post(
            `${CONFIG.BASE_URL}/api/auth/register`,
            JSON.stringify(user),
            { headers: getHeaders() }
        );
        const duration = Date.now() - startTime;

        registerDuration.add(duration);

        const success = check(res, {
            'register status is 201': (r) => r.status === 201,
            'register returns token': (r) => {
                try {
                    const body = r.json();
                    return body.data && body.data.token;
                } catch {
                    return false;
                }
            },
            'register returns user data': (r) => {
                try {
                    const body = r.json();
                    return body.data && body.data.user && body.data.user.email === user.email;
                } catch {
                    return false;
                }
            },
        });

        registerSuccessRate.add(success);
        if (success) {
            registeredUsers.add(1);
        }

        sleep(Math.random() * 2 + 1); // 1-3 seconds between registrations
    });
}

// Scenario 2: Login Test
export function loginScenario(data) {
    group('User Login', () => {
        const user = data.user;

        const startTime = Date.now();
        const res = http.post(
            `${CONFIG.BASE_URL}/api/auth/login`,
            JSON.stringify({
                email: user.email,
                password: user.password,
            }),
            { headers: getHeaders() }
        );
        const duration = Date.now() - startTime;

        loginDuration.add(duration);

        const success = check(res, {
            'login status is 200': (r) => r.status === 200,
            'login returns token': (r) => {
                try {
                    const body = r.json();
                    return body.data && body.data.token;
                } catch {
                    return false;
                }
            },
            'login returns user data': (r) => {
                try {
                    const body = r.json();
                    return body.data && body.data.user;
                } catch {
                    return false;
                }
            },
        });

        loginSuccessRate.add(success);
        if (success) {
            successfulLogins.add(1);
        }

        sleep(Math.random() + 0.5); // 0.5-1.5 seconds between logins
    });
}

// Default function (runs if no specific scenario)
export default function (data) {
    loginScenario(data);
}
