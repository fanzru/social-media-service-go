import http from "k6/http";
import { check, sleep } from "k6";

// Test configuration
export const options = {
  stages: [
    { duration: "30s", target: 10 }, // Ramp up to 10 users over 30 seconds
    { duration: "2m", target: 10 }, // Stay at 10 users for 2 minutes (100 RPM)
    { duration: "30s", target: 0 }, // Ramp down to 0 users over 30 seconds
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
    http_req_failed: ["rate<0.1"], // Error rate should be less than 10%
  },
};

// Base URL
const BASE_URL = "http://localhost:8080";

// Test scenarios
const scenarios = [
  {
    name: "GET /api/posts",
    url: `${BASE_URL}/api/posts?limit=20`,
    method: "GET",
    weight: 60, // 60% of requests
  },
  {
    name: "GET /api/account/profile",
    url: `${BASE_URL}/api/account/profile`,
    method: "GET",
    weight: 20, // 20% of requests
  },
  {
    name: "GET /api/comments",
    url: `${BASE_URL}/api/comments`,
    method: "GET",
    weight: 20, // 20% of requests
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
      },
    }
  );

  // Check response based on scenario
  const checks = {
    "status is not 500": (r) => r.status !== 500,
    "response time < 1000ms": (r) => r.timings.duration < 1000,
  };

  // Add specific checks based on endpoint
  if (selectedScenario.name === "GET /api/posts") {
    checks["status is 200"] = (r) => r.status === 200;
    checks["response has posts data"] = (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.code === "SUCCESS" && body.data && body.data.posts;
      } catch (e) {
        return false;
      }
    };
  } else if (selectedScenario.name === "GET /api/account/profile") {
    checks["status is 401 or 200"] = (r) =>
      r.status === 401 || r.status === 200;
    checks["response has proper structure"] = (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.code && body.message;
      } catch (e) {
        return false;
      }
    };
  } else if (selectedScenario.name === "GET /api/comments") {
    checks["status is 404 or 200"] = (r) =>
      r.status === 404 || r.status === 200;
  }

  check(response, checks);

  // Wait between requests (to achieve ~100 RPM)
  sleep(0.6); // 60 seconds / 100 requests = 0.6 seconds per request
}

export function handleSummary(data) {
  return {
    "k6-results.json": JSON.stringify(data, null, 2),
    stdout: `
========================================
K6 Load Test Results Summary
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
- GET /api/posts: 60% of requests
- GET /api/account/profile: 20% of requests  
- GET /api/comments: 20% of requests
========================================
    `,
  };
}
