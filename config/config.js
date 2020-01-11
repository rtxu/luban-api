export const GITHUB_OAUTH_APP = {
  client_id:
    process.env.GITHUB_OAUTH_APP_CLIENT_ID ||
    "<your_github_oauth_app_client_id>",
  client_secret:
    process.env.GITHUB_OAUTH_APP_CLIENT_SECRET ||
    "<your_github_oauth_app_client_secret>"
};

export const JWT_SECRET = process.env.JWT_SECRET || "<your_secret_to_sign_gwt>";
export const APP_ROOT = process.env.APP_ROOT || "<your_web_app_root>";
