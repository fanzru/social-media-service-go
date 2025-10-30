import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate } from "k6/metrics";
import { SharedArray } from "k6/data";

// Config via env
const SERVER_PORT = __ENV.SERVER_PORT || "8080";
const BASE_URL = `http://localhost:${SERVER_PORT}`;
const AUTH_TOKEN = __ENV.AUTH_TOKEN || "";
const SAMPLE_IMAGE = __ENV.SAMPLE_IMAGE || "";
const POST_ID_FOR_COMMENTS = __ENV.POST_ID || "1";

// Targets per requirement
// - 1k uploaded images per hour ≈ 0.278 req/s
// - 100k new comments per hour ≈ 27.78 req/s

export const options = {
  scenarios: {
    // Upload scenario is enabled only if SAMPLE_IMAGE is provided
    uploads: SAMPLE_IMAGE
      ? {
          executor: "constant-arrival-rate",
          rate: 1000, // per timeUnit
          timeUnit: "1h",
          duration: "1h",
          preAllocatedVUs: 10,
          maxVUs: 50,
          exec: "uploadImage",
        }
      : undefined,
    comments: {
      executor: "constant-arrival-rate",
      rate: 100000, // per timeUnit
      timeUnit: "1h",
      duration: "1h",
      preAllocatedVUs: 100,
      maxVUs: 500,
      exec: "createComment",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<50"], // non-upload endpoints ideally
  },
};

const uploadTrend = new Trend("upload_duration");
const commentTrend = new Trend("comment_duration");
const successRate = new Rate("success_rate");

export function setup() {
  if (!AUTH_TOKEN) {
    // You can export a token via AUTH_TOKEN env to hit protected endpoints
    console.warn(
      "AUTH_TOKEN not set. Requests may fail for protected endpoints."
    );
  }
  if (!SAMPLE_IMAGE) {
    console.warn("SAMPLE_IMAGE not set. Upload scenario will be disabled.");
  }
}

export function uploadImage() {
  if (!SAMPLE_IMAGE) {
    // Should not happen when scenario is disabled; guard anyway
    sleep(1);
    return;
  }

  // k6 requires the file to be available via open() from the working directory
  // Pass SAMPLE_IMAGE env as a relative path, e.g., scripts/k6/sample.jpg
  const bin = open(SAMPLE_IMAGE, "b");
  const form = {
    caption: `k6 upload at ${new Date().toISOString()}`,
    image: http.file(bin, "upload.jpg", "image/jpeg"),
  };

  const params = {
    headers: AUTH_TOKEN ? { Authorization: `Bearer ${AUTH_TOKEN}` } : {},
  };

  const res = http.post(`${BASE_URL}/api/posts`, form, params);
  uploadTrend.add(res.timings.duration);
  successRate.add(res.status >= 200 && res.status < 400);
  check(res, {
    "upload status is 201/200": (r) => r.status === 201 || r.status === 200,
  });
}

export function createComment() {
  const url = `${BASE_URL}/api/comments`;
  const payload = JSON.stringify({
    post_id: Number(POST_ID_FOR_COMMENTS),
    content: `k6 comment at ${new Date().toISOString()}`,
  });
  const params = {
    headers: {
      "Content-Type": "application/json",
      ...(AUTH_TOKEN ? { Authorization: `Bearer ${AUTH_TOKEN}` } : {}),
    },
  };
  const res = http.post(url, payload, params);
  commentTrend.add(res.timings.duration);
  successRate.add(res.status >= 200 && res.status < 400);
  check(res, {
    "comment status is 201/200": (r) => r.status === 201 || r.status === 200,
  });
}

export default function () {
  // Not used; scenarios call exec functions directly
}
