import jwt from "jsonwebtoken";

import { GITHUB_OAUTH_APP, JWT_SECRET, APP_ROOT } from "../../config/config";
import { request, makeQueryString } from "../utils";

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

export default async (req, res) => {
  try {
    const userInfo = await fetchUserInfoFromGithub(req, res);

    // query db
    // user not exist => insert db
    // user exist

    loginSuccess(userInfo.login, res);
  } catch (e) {
    console.error("caught exception: ", e.name, e.message);
  }
};
