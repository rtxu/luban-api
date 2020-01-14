import express from "express";

import githubLoginCallback from "./src/handlers/github_login_callback";

const app = express();
const port = 3000;

/**
 * Common Middleware
 */

/**
 * Business Logic
 */
app.get(
  ["login", "signup"].map(action => `/callback/github/${action}`),
  githubLoginCallback
);

/**
 * Error Handling Middleware
 */

app.listen(port, () => console.log(`Example app listening on port ${port}!`));
