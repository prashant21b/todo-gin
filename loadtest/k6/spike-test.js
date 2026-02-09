/**
 * K6 Spike Test - Sudden Traffic Burst
 * Tests how the system handles sudden increases in traffic
 *
 * Run: k6 run loadtest/k6/spike-test.js
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';
import { CONFIG, generateTestUser, generateTodo, getHeaders } from './config.js';

// Custom metrics
const spikeSuccess = new Rate('spike_success');
const responseDuring = new Trend('response_during_spike', true);

export const options = {
    scenarios: {
        spike: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '30s', target: 10 },    // Normal load
                { duration: '1m', target: 10 },     // Stay normal
                { duration: '10s', target: 100 },   // SPIKE! Rapid increase
                { duration: '1m', target: 100 },    // Stay at spike
                { duration: '10s', target: 10 },    // Rapid decrease
                { duration: '1m', target: 10 },     // Recovery
                { duration: '30s', target: 0 },     // Ramp down
            ],
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<2000'],  // Allow higher latency during spike
        http_req_failed: ['rate<0.10'],      // Allow up to 10% failure during spike
        spike_success: ['rate>0.85'],
    },
};

// Setup
export function setup() {
    const user = generateTestUser();

    const res = http.post(
        `${CONFIG.BASE_URL}/api/auth/register`,
        JSON.stringify(user),
        { headers: getHeaders() }
    );

    if (res.status !== 201) {
        console.error('Setup failed');
        return { token: null };
    }

    return {
        token: res.json('data.token'),
        user,
    };
}

// Main test - Simulates user activity during spike
export default function (data) {
    if (!data.token) return;

    const headers = getHeaders(data.token);
    let success = true;

    // Quick CRUD cycle
    const todo = generateTodo();

    // Create
    const startTime = Date.now();
    const createRes = http.post(
        `${CONFIG.BASE_URL}/api/todos`,
        JSON.stringify(todo),
        { headers }
    );

    if (!check(createRes, { 'create ok': (r) => r.status === 201 })) {
        success = false;
    }

    const todoId = createRes.json('data.id');

    // Read
    if (todoId) {
        const getRes = http.get(`${CONFIG.BASE_URL}/api/todos/${todoId}`, { headers });
        if (!check(getRes, { 'get ok': (r) => r.status === 200 })) {
            success = false;
        }

        // Delete
        const delRes = http.del(`${CONFIG.BASE_URL}/api/todos/${todoId}`, null, { headers });
        if (!check(delRes, { 'delete ok': (r) => r.status === 200 })) {
            success = false;
        }
    }

    const duration = Date.now() - startTime;
    responseDuring.add(duration);
    spikeSuccess.add(success);

    sleep(Math.random() * 0.5);
}

export function teardown(data) {
    console.log('\n========================================');
    console.log('Spike Test Completed');
    console.log('Check if system recovered properly after spike');
    console.log('========================================\n');
}
