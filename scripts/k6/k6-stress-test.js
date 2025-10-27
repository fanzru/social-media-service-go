import http from "k6/http";
import { check, sleep } from "k6";

// Stress test configuration
export const options = {
  stages: [
    { duration: "1m", target: 20 }, // Ramp up to 20 users over 1 minute
    { duration: "2m", target: 50 }, // Ramp up to 50 users over 2 minutes
    { duration: "3m", target: 50 }, // Stay at 50 users for 3 minutes
    { duration: "1m", target: 100 }, // Spike to 100 users for 1 minute
    { duration: "2m", target: 50 }, // Ramp down to 50 users
    { duration: "1m", target: 0 }, // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ["p(95)<1000"], // 95% of requests should be below 1000ms
    http_req_failed: ["rate<0.2"], // Error rate should be less than 20%
  },
};

// Base URL
const BASE_URL = "http://localhost:8080";

// Test scenarios for stress test
const scenarios = [
  {
    name: "GET /api/posts",
    url: `${BASE_URL}/api/posts?limit=20`,
    method: "GET",
    weight: 50, // 50% of requests
  },
  {
    name: "GET /api/posts with pagination",
    url: `${BASE_URL}/api/posts?limit=10&offset=0`,
    method: "GET",
    weight: 20, // 20% of requests
  },
  {
    name: "GET /api/account/profile",
    url: `${BASE_URL}/api/account/profile`,
    method: "GET",
    weight: 15, // 15% of requests
  },
  {
    name: "GET /api/comments",
    url: `${BASE_URL}/api/comments`,
    method: "GET",
    weight: 10, // 10% of requests
  },
  {
    name: "GET /health",
    url: `${BASE_URL}/health`,
    method: "GET",
    weight: 5, // 5% of requests
  },
];

export default function () {
  // Select scenario based on weight
  const random = Math.random() * 100;
  let cumulativeWeight = 0;
  let selectedScenario = scenarios[0];

  for (const scenario of scenarios) {
    cumulativeWeight += scenario.weight;
    if (random <= cumulativeWeight) {
      selectedScenario = scenario;
      break;
    }
  }

  // Execute the selected scenario
  const response = http.request(
    selectedScenario.method,
    selectedScenario.url,
    null,
    {
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        "User-Agent": "K6-StressTest/1.0",
      },
    }
  );

  // Basic checks for all scenarios
  const checks = {
    "status is not 500": (r) => r.status !== 500,
    "response time < 2000ms": (r) => r.timings.duration < 2000,
    "response has body": (r) => r.body && r.body.length > 0,
  };

  // Add specific checks based on endpoint
  if (selectedScenario.name.includes("/api/posts")) {
    checks["status is 200"] = (r) => r.status === 200;
    checks["response has posts data"] = (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.code === "SUCCESS" && body.data;
      } catch (e) {
        return false;
      }
    };
  } else if (selectedScenario.name.includes("/api/account/profile")) {
    checks["status is 401 or 200"] = (r) =>
      r.status === 401 || r.status === 200;
  } else if (selectedScenario.name.includes("/health")) {
    checks["status is 200"] = (r) => r.status === 200;
    checks["response has health data"] = (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.status === "healthy";
      } catch (e) {
        return false;
      }
    };
  }

  check(response, checks);

  // Variable sleep time based on load
  const sleepTime = Math.random() * 0.5 + 0.1; // Random sleep between 0.1-0.6 seconds
  sleep(sleepTime);
}

export function handleSummary(data) {
  return {
    "k6-stress-results.json": JSON.stringify(data, null, 2),
    stdout: `
========================================
K6 Stress Test Results Summary
========================================
Total Requests: ${data.metrics.http_reqs.values.count}
Failed Requests: ${data.metrics.http_req_failed.values.count}
Success Rate: ${(
      (1 -
        data.metrics.http_req_failed.values.count /
          data.metrics.http_reqs.values.count) *
      100
    ).toFixed(2)}%

Response Times:
- Average: ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms
- Median: ${data.metrics.http_req_duration.values.med.toFixed(2)}ms
- 95th percentile: ${data.metrics.http_req_duration.values["p(95)"].toFixed(
      2
    )}ms
- 99th percentile: ${data.metrics.http_req_duration.values["p(99)"].toFixed(
      2
    )}ms

Requests per second: ${data.metrics.http_reqs.values.rate.toFixed(2)}

Test Scenarios:
- GET /api/posts: 50% of requests
- GET /api/posts (pagination): 20% of requests
- GET /api/account/profile: 15% of requests
- GET /api/comments: 10% of requests
- GET /health: 5% of requests

Load Pattern:
- 0-1m: Ramp up to 20 users
- 1-3m: Ramp up to 50 users
- 3-6m: Stay at 50 users
- 6-7m: Spike to 100 users
- 7-9m: Ramp down to 50 users
- 9-10m: Ramp down to 0 users
========================================
    `,
  };
}
