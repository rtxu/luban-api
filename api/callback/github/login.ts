import { NowRequest, NowResponse } from "@now/node";

import expressAdapter from "../../_middleware/express_adapter";
import githubLoginCallback from "../../../src/handlers/github_login_callback";

export default function(req: NowRequest, res: NowResponse) {
  expressAdapter(req, res);
  githubLoginCallback(req, res);
}
