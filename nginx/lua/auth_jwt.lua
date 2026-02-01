local jwt = require("resty.jwt")

local JWT_SECRET = os.getenv("JWT_SECRET")

if not JWT_SECRET then
    ngx.log(ngx.ERR, "JWT_SECRET is not defined!")
    return ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
end

local auth_header = ngx.var.http_Authorization
if not auth_header then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.say("Missing Authorization header!")
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

local _, _, token = string.find(auth_header, "Bearer%s+(.+)")
if not token then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.say("Invalid Authorization header!")
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

local jwt_obj = jwt:verify(JWT_SECRET, token)
if not jwt_obj["verified"] then
    ngx.status = ngx.HTTP_UNAUTHORIZED
    ngx.say("Invalid token: ", jwt_obj.reason)
    -- Use refresh token!
    return ngx.exit(ngx.HTTP_UNAUTHORIZED)
end

ngx.req.set_header("X-User-ID", jwt_obj.payload.sub)