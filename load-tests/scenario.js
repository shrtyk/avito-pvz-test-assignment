import http from 'k6/http';
import { check, sleep, group } from 'k6';

export const options = {
  scenarios: {
    contacts: {
      executor: 'ramping-vus',
      startVUs: 50,
      stages: [
        { duration: '30s', target: 1000 }, // Ramp up to 1000 VUs
        { duration: '1m', target: 1000 },  // Stay at 1000 VUs for 1 minute
        { duration: '10s', target: 0 },   // Ramp down
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: {
    'http_req_duration': ['p(95)<100'],
    'checks': ['rate>0.9999'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// The Test Scenario
export default function () {
  // Step 1: get a moderator token and create a unique PVZ
  let loginRes = http.post(`${BASE_URL}/dummyLogin`, JSON.stringify({ role: 'moderator' }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: '/dummyLogin (moderator)' },
  });
  check(loginRes, { 'moderator login successful': (r) => r.status === 200 });
  const moderatorToken = loginRes.json('jwt');
  const modAuthParams = { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${moderatorToken}` } };

  const pvzRes = http.post(`${BASE_URL}/pvz`, JSON.stringify({ city: 'Москва' }), {
    ...modAuthParams,
    tags: { name: '/pvz (create)' },
  });
  check(pvzRes, { 'per-VU PVZ created': (r) => r.status === 201 });
  if (pvzRes.status !== 201) {
    return; // Stop if we can't create a PVZ
  }
  const pvzId = pvzRes.json('id');

  // Step 2: get an employee token
  loginRes = http.post(`${BASE_URL}/dummyLogin`, JSON.stringify({ role: 'employee' }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: '/dummyLogin (employee)' },
  });
  check(loginRes, { 'employee login successful': (r) => r.status === 200 });
  const employeeToken = loginRes.json('jwt');
  const empAuthParams = { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${employeeToken}` } };

  group('Reception and Products Workload', () => {
    // Step 3: open a reception with the employee token
    const receptionPayload = JSON.stringify({ pvzId: pvzId });
    const receptionRes = http.post(`${BASE_URL}/receptions`, receptionPayload, {
      ...empAuthParams,
      tags: { name: '/receptions (create)' },
    });
    check(receptionRes, { 'reception opened': (r) => r.status === 201 });
    if (receptionRes.status !== 201) { return; }

    // Step 4: add 5 products to this reception
    for (let i = 0; i < 5; i++) {
      const productPayload = JSON.stringify({ pvzId: pvzId, type: 'одежда' });
      const productRes = http.post(`${BASE_URL}/products`, productPayload, {
        ...empAuthParams,
        tags: { name: '/products (create)' },
      });
      check(productRes, { 'product added': (r) => r.status === 201 });
      sleep(0.1); // Small pause between adding products
    }
  });

  // Step 5: fetching a paginated list with date filters
  group('Read PVZ Data with Filters', () => {
    const page = Math.floor(Math.random() * 10) + 1; // Random page from 1 to 10
    const limit = 10;

    const endDate = new Date();
    const startDate = new Date();
    startDate.setDate(endDate.getDate() - 7); // last 7 days

    const url = `${BASE_URL}/pvz?page=${page}&limit=${limit}&startDate=${encodeURIComponent(startDate.toISOString())}&endDate=${encodeURIComponent(endDate.toISOString())}`;

    const getPvzRes = http.get(url, {
      ...empAuthParams,
      tags: { name: '/pvz (filtered get)' },
    });
    check(getPvzRes, { 'get pvz list with filters successful': (r) => r.status === 200 });
  });

  sleep(1); // Final sleep before next iteration
}
