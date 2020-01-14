/**
 * adapt NowRequest/NowResponse into express-like Request/Response
 */

export default function(req, res) {
  res.redirect = function(status, location) {
    res.statusCode = status;
    res.setHeader("Location", location);
    res.end();
  };
}
