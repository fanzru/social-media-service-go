import http from "k6/http";
import { check, sleep } from "k6";

// Test configuration
export const options = {
  stages: [
    { duration: "30s", target: 10 }, // Ramp up to 10 users over 30 seconds
    { duration: "1m", target: 10 }, // Stay at 10 users for 1 minute (100 RPM)
    { duration: "30s", target: 0 }, // Ramp down to 0 users over 30 seconds
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"], // 95% of requests should be below 500ms
    http_req_failed: ["rate<0.1"], // Error rate should be less than 10%
  },
};

// Base URL
const BASE_URL = "http://localhost:8080";

export default function () {
  // Test GET /api/posts endpoint
  const response = http.get(`${BASE_URL}/api/posts?limit=20`, {
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
  });

  // Check response
  check(response, {
    "status is 200": (r) => r.status === 200,
    "response time < 500ms": (r) => r.timings.duration < 500,
    "response has posts data": (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.code === "SUCCESS" && body.data && body.data.posts;
      } catch (e) {
        return false;
      }
    },
  });

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
========================================
    `,
  };
}
