import http from "k6/http";
import { check, sleep } from "k6";
import { Trend, Rate } from "k6/metrics";
import { SharedArray } from "k6/data";

// Config via env
const SERVER_PORT = __ENV.SERVER_PORT || "8080";
const BASE_URL = `http://localhost:${SERVER_PORT}`;
const AUTH_TOKEN = __ENV.AUTH_TOKEN || "";
// Default to repo sample image if not provided
// Use absolute path to avoid issues when running from different directories
const SAMPLE_IMAGE =
  __ENV.SAMPLE_IMAGE ||
  "/Users/mbprom4pro/go/src/github.com/fanzru/social-media-service-go/examples/images/test.png";
const POST_ID_FOR_COMMENTS = __ENV.POST_ID || "1";
const POST_SEED_COUNT = Number(__ENV.POST_SEED_COUNT || 50);

// Load image once in init context (required by k6). If missing, uploads scenario will warn.
let IMAGE_BIN = null;
let IMAGE_MIME = "image/jpeg";
try {
  IMAGE_BIN = open(SAMPLE_IMAGE, "b");
  if (SAMPLE_IMAGE.toLowerCase().endsWith(".png")) {
    IMAGE_MIME = "image/png";
  } else if (
    SAMPLE_IMAGE.toLowerCase().endsWith(".jpg") ||
    SAMPLE_IMAGE.toLowerCase().endsWith(".jpeg")
  ) {
    IMAGE_MIME = "image/jpeg";
  }
} catch (e) {
  // Will be handled in setup/scenario enabling
}

// Targets per requirement
// - 1k uploaded images per hour ≈ 0.278 req/s
// - 100k new comments per hour ≈ 27.78 req/s

export const options = {
  scenarios: {
    // Posts scenario (multipart form: caption + image)
    // Always define executor; `createPost` will no-op if no IMAGE_BIN
    createPosts: {
      executor: "constant-arrival-rate",
      rate: 5,
      timeUnit: "1s",
      duration: "1h",
      preAllocatedVUs: 10,
      maxVUs: 50,
      exec: "createPost",
    },
    createComments: {
      executor: "constant-arrival-rate",
      rate: 100000, // per timeUnit
      timeUnit: "1h",
      duration: "1h",
      preAllocatedVUs: 100,
      maxVUs: 500,
      exec: "createComment",
    },
    getPosts: {
      executor: "constant-arrival-rate",
      rate: 10,
      timeUnit: "1s",
      duration: "1h",
      preAllocatedVUs: 10,
      maxVUs: 50,
      exec: "listPosts",
    },
    getComments: {
      executor: "constant-arrival-rate",
      rate: 10,
      timeUnit: "1s",
      duration: "1h",
      preAllocatedVUs: 10,
      maxVUs: 50,
      exec: "listComments",
    },
  },
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<50"], // non-upload endpoints ideally
  },
  // Keep setup fast; only login happens there
  setupTimeout: "10s",
};

const uploadTrend = new Trend("upload_duration");
const commentTrend = new Trend("comment_duration");
const successRate = new Rate("success_rate");

// Local post ID pool per VU; populated by successful POST /api/posts
let POST_POOL = [];

function refreshPostPool(baseUrl, token) {
  const params = token ? { headers: { Authorization: `Bearer ${token}` } } : {};
  const res = http.get(`${baseUrl}/api/posts`, params);
  if (res.status === 200) {
    try {
      const body = res.json();
      const items = body?.data || body?.posts || [];
      const ids = Array.isArray(items)
        ? items
            .map((p) => p?.id || p?.post?.id)
            .filter((v) => typeof v === "number")
        : [];
      if (ids.length > 0) {
        POST_POOL = ids;
      }
    } catch (e) {
      // ignore parse errors
    }
  }
}

export function setup() {
  let token = AUTH_TOKEN;
  let postId = Number(POST_ID_FOR_COMMENTS) || 0;

  if (!token) {
    const loginPayload = JSON.stringify({
      email: "john@example.com",
      password: "password123",
    });
    const params = { headers: { "Content-Type": "application/json" } };
    const res = http.post(
      `${BASE_URL}/api/account/login`,
      loginPayload,
      params
    );

    if (res.status === 200) {
      try {
        const body = res.json();
        // API uses StandardResponse with data.access_token
        token = body?.data?.access_token || "";
        if (!token) {
          console.warn("Login succeeded but access_token missing in response.");
        }
      } catch (e) {
        console.error("Failed to parse login response:", e);
      }
    } else {
      console.warn(`Login failed with status ${res.status}. Using no token.`);
    }
  }

  if (!SAMPLE_IMAGE) {
    console.warn("SAMPLE_IMAGE not set. Upload scenario will be disabled.");
  }

  // Only return token and optional static postId; no seeding in setup
  return { token, postId };
}

export function createPost(data) {
  if (!SAMPLE_IMAGE || !IMAGE_BIN) {
    // Should not happen when scenario is disabled; guard anyway
    sleep(1);
    return;
  }

  const form = {
    caption: `k6 upload at ${new Date().toISOString()}`,
    image: http.file(
      IMAGE_BIN,
      "upload" + (IMAGE_MIME === "image/png" ? ".png" : ".jpg"),
      IMAGE_MIME
    ),
  };

  const bearer = data?.token || AUTH_TOKEN || "";
  const params = {
    headers: bearer ? { Authorization: `Bearer ${bearer}` } : {},
  };

  const res = http.post(`${BASE_URL}/api/posts`, form, params);
  uploadTrend.add(res.timings.duration);
  successRate.add(res.status >= 200 && res.status < 400);
  // On success, capture ID into local POST_POOL for later comments
  if (res.status === 201 || res.status === 200) {
    try {
      const body = res.json();
      const id = body?.data?.id || body?.data?.post?.id;
      if (id) POST_POOL.push(Number(id));
    } catch (e) {
      // ignore parse errors
    }
  }
  check(res, {
    "upload status is 201/200": (r) => r.status === 201 || r.status === 200,
  });
}

export function createComment(data) {
  const bearer = data?.token || AUTH_TOKEN || "";
  let targetPostId = null;

  // Always create a post first and use its ID for commenting (ties comments to real created posts)
  if (bearer && IMAGE_BIN) {
    const form = {
      caption: `seed for comment at ${new Date().toISOString()}`,
      image: http.file(
        IMAGE_BIN,
        "seed" + (IMAGE_MIME === "image/png" ? ".png" : ".jpg"),
        IMAGE_MIME
      ),
    };
    const postParams = { headers: { Authorization: `Bearer ${bearer}` } };
    const postRes = http.post(`${BASE_URL}/api/posts`, form, postParams);
    uploadTrend.add(postRes.timings.duration);
    successRate.add(postRes.status >= 200 && postRes.status < 400);

    // Extract the new post ID from response and add to pool
    if (postRes.status === 201 || postRes.status === 200) {
      try {
        const body = postRes.json();
        const newId = body?.data?.id;
        if (newId) {
          targetPostId = Number(newId);
          POST_POOL.push(targetPostId);
        }
      } catch (e) {
        // ignore parse errors
      }
    }
  }

  // If we didn't just create a post, try local pool or refresh from API
  if (!targetPostId) {
    if (POST_POOL.length > 0) {
      targetPostId = POST_POOL[Math.floor(Math.random() * POST_POOL.length)];
    } else {
      refreshPostPool(BASE_URL, bearer);
      if (POST_POOL.length > 0) {
        targetPostId = POST_POOL[Math.floor(Math.random() * POST_POOL.length)];
      }
    }
  }

  if (!targetPostId) {
    // No post id available; skip to avoid 404
    sleep(0.05);
    return;
  }

  // Now create comment on the target post
  const url = `${BASE_URL}/api/comments/by-post/${targetPostId}`;
  const payload = JSON.stringify({
    content: `k6 comment at ${new Date().toISOString()}`,
  });
  const commentParams = {
    headers: {
      "Content-Type": "application/json",
      ...(bearer ? { Authorization: `Bearer ${bearer}` } : {}),
    },
  };
  const res = http.post(url, payload, commentParams);
  commentTrend.add(res.timings.duration);
  successRate.add(res.status >= 200 && res.status < 400);
  check(res, {
    "comment status is 201/200": (r) => r.status === 201 || r.status === 200,
  });
}

export default function () {
  // Not used; scenarios call exec functions directly
}

export function listPosts(data) {
  const res = http.get(`${BASE_URL}/api/posts`);
  check(res, { "list posts 200": (r) => r.status === 200 });
}

export function listComments(data) {
  const pid = data?.postId || Number(POST_ID_FOR_COMMENTS);
  const res = http.get(`${BASE_URL}/api/comments/by-post/${pid}`);
  check(res, { "list comments 200": (r) => r.status === 200 });
}
