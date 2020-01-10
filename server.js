import express from "express";
import jwt from "jsonwebtoken";

import { GITHUB_OAUTH_APP, JWT_SECRET, APP_ROOT } from "./config/config";
import { request, makeQueryString } from "./util";

const app = express();
const port = 3000;

/**
 * Common Middleware
 */

/**
 * Business Logic
 */
async function fetchUserInfoFromGithub(req, res) {
  const code = req.query.code;
  const params = {
    ...GITHUB_OAUTH_APP,
    code
  };
  const query = makeQueryString(params);
  const githubAccessToken = await request(
    `https://github.com/login/oauth/access_token?${query}`,
    {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json"
      }
    }
  );

  const userJson = await request("https://api.github.com/user", {
    method: "GET",
    headers: {
      Authorization: `${githubAccessToken.token_type} ${githubAccessToken.access_token}`
    }
  });

  return userJson;
}

function loginSuccess(uid, res) {
  // sign token
  const jwtToken = jwt.sign({ uid }, JWT_SECRET, {
    expiresIn: "7d"
  });
  const params = {
    access_token: `Bearer ${jwtToken}`
  };
  const query = makeQueryString(params);
  res.redirect(303, `${APP_ROOT}/login-success?${query}`);
}

app.get(
  ["login", "signup"].map(action => `/callback/github/${action}`),
  async (req, res, next) => {
    try {
      const userInfo = await fetchUserInfoFromGithub(req, res);

      // query db
      // user not exist => insert db
      // user exist

      loginSuccess(userInfo.login, res);
    } catch (e) {
      next(e);
    }
  }
);

/**
 * Error Handling Middleware
 */

app.listen(port, () => console.log(`Example app listening on port ${port}!`));
